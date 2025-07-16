package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather/internal/adapter/benchmark"
	"weather/internal/adapter/cache"
	"weather/internal/delivery/grpcapi"
	"weather/internal/delivery/httpapi"
	weatherpb "weather/internal/proto"

	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc"

	"weather/internal/app/di"
	"weather/internal/config"
	"weather/internal/infra"
	"weather/internal/service"

	"github.com/redis/go-redis/v9"
)

type App struct {
	HttpServer *http.Server
	HttpLis    net.Listener
	Redis      *redis.Client
	Logger     *log.Logger
	GrpcServer *grpc.Server
	GrpcLis    net.Listener
}

func Run(logger *log.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}

	if err := app.Start(); err != nil {
		logger.Printf("Starting application: %v", err)
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Println("Shutdown signal received")
	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Shutdown(shutdownCtx)
}

func NewApp(ctx context.Context, cfg *config.Config, logger *log.Logger) (*App, error) {
	var redisClient *redis.Client
	var metrics di.CacheMetrics
	var httpClient *http.Client
	var weatherProvider service.Provider

	if cfg.BenchmarkMode {
		logger.Println(" Running in BENCHMARK MODE — skipping Redis and using BenchmarkProvider")
		redisClient = nil
		metrics = cache.NewNoopMetrics()
		weatherProvider = benchmark.NewProvider()

	} else {
		redis, err := infra.NewRedisClient(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("redis error: %w", err)
		}
		redisClient = redis

		metrics = cache.NewMetrics()
		metrics.Register()

		httpClient = &http.Client{Timeout: 5 * time.Second}

		weatherProvider = di.BuildProviders(di.ProviderDeps{
			Cfg:         cfg,
			Logger:      logger,
			RedisClient: redisClient,
			HttpClient:  httpClient,
			Metrics:     metrics,
		})
	}

	weatherService := service.NewService(weatherProvider)

	// HTTP
	mux := http.NewServeMux()
	weatherHandler := httpapi.NewWeatherHandler(weatherService)
	httpapi.RegisterRoutes(mux, weatherHandler)

	httpLis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("listen HTTP: %w", err)
	}

	httpServer := &http.Server{Handler: mux}

	// gRPC
	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("listen gRPC: %w", err)
	}

	grpcServer := grpc.NewServer()
	grpcHandler := grpcapi.NewHandler(weatherService)
	weatherpb.RegisterWeatherServiceServer(grpcServer, grpcHandler)

	logger.Printf("[INFO] Application initialized successfully")
	return &App{
		HttpServer: httpServer,
		HttpLis:    httpLis,
		GrpcServer: grpcServer,
		GrpcLis:    grpcLis,
		Redis:      redisClient,
		Logger:     logger,
	}, nil
}

func (a *App) Start() error {
	a.Logger.Println("[INFO] Starting servers...")

	var g errgroup.Group

	g.Go(func() error {
		a.Logger.Printf("[INFO] HTTP server starting on %s", a.HttpLis.Addr())
		if err := a.HttpServer.Serve(a.HttpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		a.Logger.Printf("[INFO] gRPC server starting on %s", a.GrpcLis.Addr())
		if err := a.GrpcServer.Serve(a.GrpcLis); err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	go func() {
		if err := g.Wait(); err != nil {
			a.Logger.Printf("[ERROR] Server exited with error: %v", err)
		}
	}()

	a.Logger.Println("[INFO] All servers started successfully")
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Println("[INFO] Initiating graceful shutdown...")

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Printf("[ERROR] Redis close error: %v", err)
		}
	}

	if err := a.HttpServer.Shutdown(ctx); err != nil {
		a.Logger.Printf("[ERROR] HTTP shutdown error: %v", err)
	} else {
		a.Logger.Printf("[INFO] HTTP server stopped")
	}

	done := make(chan struct{})
	go func() {
		a.GrpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		a.Logger.Printf("[INFO] gRPC server stopped gracefully")
	case <-ctx.Done():
		a.Logger.Printf("[WARN] gRPC server force stopped due to timeout")
		a.GrpcServer.Stop()
	}
	return nil
}
