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

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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

	ctx = loggerPkg.With(ctx, lg)

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}

	if err := app.Start(ctx); err != nil {
		logger := loggerPkg.From(ctx)
		logger.Errorf("Starting application: %v", err)
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger := loggerPkg.From(ctx)
	logger.Info("Shutdown signal received")
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

	logger := loggerPkg.From(ctx)

	if cfg.BenchmarkMode {
		logger.Info(" Running in BENCHMARK MODE — skipping Redis and using BenchmarkProvider")
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
			RedisClient: redisClient,
			HttpClient:  httpClient,
			Metrics:     metrics,
		})
	}

	weatherService := weather.NewService(weatherProvider)

	// HTTP
	mux := http.NewServeMux()
	weatherHandler := httpapi.NewHandler(weatherService)
	httpapi.RegisterRoutes(mux, weatherHandler, logger)

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

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcapi.LoggingUnaryServerInterceptor(logger)),
	)
	grpcHandler := grpcapi.NewHandler(weatherService)
	weatherpb.RegisterWeatherServiceServer(grpcServer, grpcHandler)

	logger.Infof("[INFO] Application initialized successfully")
	return &App{
		HttpServer: httpServer,
		HttpLis:    httpLis,
		GrpcServer: grpcServer,
		GrpcLis:    grpcLis,
		Redis:      redisClient,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	logger := loggerPkg.From(ctx)
	logger.Infow("Starting HTTP and gRPC servers")

	var g errgroup.Group

	g.Go(func() error {
		logger.Infow("HTTP server listening", "address", a.HttpLis.Addr().String())
		if err := a.HttpServer.Serve(a.HttpLis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		logger.Infow("gRPC server listening", "address", a.GrpcLis.Addr().String())
		if err := a.GrpcServer.Serve(a.GrpcLis); err != nil {
			return fmt.Errorf("gRPC server error: %w", err)
		}
		return nil
	})

	time.Sleep(100 * time.Millisecond)

	go func() {
		if err := g.Wait(); err != nil {
			logger.Errorw("Server exited with error", "error", err)
		}
	}()

	logger.Infow("All servers started successfully")
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	logger := loggerPkg.From(ctx)
	logger.Infow("Initiating graceful shutdown")

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			logger.Errorw("Failed to close Redis connection", "error", err)
		} else {
			logger.Infow("Redis connection closed")
		}
	}

	if err := a.HttpServer.Shutdown(ctx); err != nil {
		logger.Errorw("HTTP server shutdown error", "error", err)
	} else {
		logger.Infow("HTTP server stopped")
	}

	done := make(chan struct{})
	go func() {
		a.GrpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		logger.Infow("gRPC server stopped")
	case <-ctx.Done():
		logger.Warnw("gRPC server stop timed out")
	}

	logger.Infow("Graceful shutdown completed")
	return nil
}
