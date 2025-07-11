package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email/internal/adapter/sendgrid"
	"email/internal/adapter/template"
	"email/internal/config"
	"email/internal/delivery"
	"email/internal/service"
)

type App struct {
	Server *http.Server
	Logger *log.Logger
}

func Run(logger *log.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return err
	}

	app.Start()

	// Graceful shutdown on SIGINT/SIGTERM
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
	templateStore, err := template.Load("template")
	if err != nil {
		return nil, err
	}

	emailProvider := sendgrid.New(cfg.SendGridKey, cfg.EmailFrom)
	emailService := service.NewService(emailProvider, templateStore)

	handler := delivery.NewEmailHandler(emailService, logger)

	mux := http.NewServeMux()
	delivery.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	return &App{
		Server: server,
		Logger: logger,
	}, nil
}

func (a *App) Start() {
	go func() {
		a.Logger.Printf("Email service running at %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Fatalf("server error: %v", err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Println("Shutting down email service...")

	if err := a.Server.Shutdown(ctx); err != nil {
		return err
	}

	a.Logger.Println("Shutdown complete")
	return nil
}
