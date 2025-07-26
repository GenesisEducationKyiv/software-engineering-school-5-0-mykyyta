package app

import (
	"context"
	"errors"
	"fmt"
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
	"weather/internal/weather"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	lgCtx "weather/pkg/logger"
)

type App struct {
	HttpServer *http.Server
	HttpLis    net.Listener
	Redis      *redis.Client
	GrpcServer *grpc.Server
	GrpcLis    net.Listener
}

func Run(lg *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = lgCtx.With(ctx, lg)

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}

	if err := app.Start(ctx); err != nil {
		lgCtx.From(ctx).Errorf("Starting application: %v", err)
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	lgCtx.From(ctx).Info("Shutdown signal received")
	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Shutdown(shutdownCtx)
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	var redisClient *redis.Client
	var metrics di.CacheMetrics
	var httpClient *http.Client
	var weatherProvider weather.Provider

	lg := lgCtx.From(ctx)

	if cfg.BenchmarkMode {
		lg.Info(" Running in BENCHMARK MODE — skipping Redis and using BenchmarkProvider")
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
			Logger:      lg,
			RedisClient: redisClient,
			HttpClient:  httpClient,
			Metrics:     metrics,
		})
	}

	weatherService := weather.NewService(weatherProvider)

	// HTTP
	mux := http.NewServeMux()
	weatherHandler := httpapi.NewHandler(weatherService)
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

	lg.Infof("[INFO] Application initialized successfully")
	return &App{
		HttpServer: httpServer,
		HttpLis:    httpLis,
		GrpcServer: grpcServer,
		GrpcLis:    grpcLis,
		Redis:      redisClient,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	lg := lgCtx.From(ctx)
	lg.Infow("Starting HTTP and gRPC servers")

	var g errgroup.Group

	g.Go(func() error {
		lg.Infow("HTTP server listening", "address", a.HttpLis.Addr().String())
		if err := a.HttpServer.Serve(a.HttpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		lg.Infow("gRPC server listening", "address", a.GrpcLis.Addr().String())
		if err := a.GrpcServer.Serve(a.GrpcLis); err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	go func() {
		if err := g.Wait(); err != nil {
			lg.Errorw("Server exited with error", "error", err)
		}
	}()

	lg.Infow("All servers started successfully")
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	lg := lgCtx.From(ctx)
	lg.Infow("Initiating graceful shutdown")

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			lg.Errorw("Failed to close Redis connection", "error", err)
		} else {
			lg.Infow("Redis connection closed")
		}
	}

	if err := a.HttpServer.Shutdown(ctx); err != nil {
		lg.Errorw("HTTP server shutdown error", "error", err)
	} else {
		lg.Infow("HTTP server stopped")
	}

	done := make(chan struct{})
	go func() {
		a.GrpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		lg.Infow("gRPC server stopped gracefully")
	case <-ctx.Done():
		lg.Warnw("gRPC server force stopped due to timeout")
		a.GrpcServer.Stop()
	}

	lg.Infow("Shutdown complete")
	return nil
}
