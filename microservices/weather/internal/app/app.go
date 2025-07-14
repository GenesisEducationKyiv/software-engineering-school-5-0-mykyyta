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
	"weather/internal/adapter/cache"
	"weather/internal/delivery/grpcapi"
	"weather/internal/delivery/httpapi"
	weatherpb "weather/internal/proto"

	"google.golang.org/grpc"

	"weather/internal/app/di"
	"weather/internal/config"
	"weather/internal/infra"
	"weather/internal/service"

	"github.com/redis/go-redis/v9"
)

type App struct {
	HttpServer *http.Server
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
	defer func() {
		if err := app.Redis.Close(); err != nil {
			logger.Printf("close Redis: %v", err)
		}
	}()

	app.Start()

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
	redisClient, err := infra.NewRedisClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	metrics := cache.NewMetrics()
	metrics.Register()

	httpClient := &http.Client{Timeout: 5 * time.Second}

	weatherProvider := di.BuildProviders(di.ProviderDeps{
		Cfg:         cfg,
		Logger:      logger,
		RedisClient: redisClient,
		HttpClient:  httpClient,
		Metrics:     metrics,
	})

	weatherService := service.NewService(weatherProvider)

	// HTTP
	mux := http.NewServeMux()
	weatherHandler := httpapi.NewWeatherHandler(weatherService)
	httpapi.RegisterRoutes(mux, weatherHandler)

	httpServer := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// gRPC
	grpcLis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		return nil, fmt.Errorf("listen gRPC: %w", err)
	}

	grpcServer := grpc.NewServer()
	grpcHandler := grpcapi.NewHandler(weatherService)
	weatherpb.RegisterWeatherServiceServer(grpcServer, grpcHandler)

	return &App{
		HttpServer: httpServer,
		GrpcServer: grpcServer,
		GrpcLis:    grpcLis,
		Redis:      redisClient,
		Logger:     logger,
	}, nil
}

func (a *App) Start() {
	// HTTP
	go func() {
		a.Logger.Printf("HTTP server running at %s", a.HttpServer.Addr)
		if err := a.HttpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Fatalf("HTTP server error: %v", err)
		}
	}()

	// gRPC
	go func() {
		a.Logger.Println("gRPC server running at :50051")
		if err := a.GrpcServer.Serve(a.GrpcLis); err != nil {
			a.Logger.Fatalf("gRPC server error: %v", err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Printf("Redis close error: %v", err)
		}
	}

	if err := a.HttpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP server shutdown error: %w", err)
	}

	a.GrpcServer.GracefulStop()

	log.Println("Shutdown complete")
	return nil
}
