package job

import (
	"context"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"github.com/robfig/cron/v3"
)

type CronEventSource struct {
	cron   *cron.Cron
	events chan string
}

func NewCronEventSource() *CronEventSource {
	return &CronEventSource{
		cron:   cron.New(),
		events: make(chan string, 10),
	}
}

func (s *CronEventSource) Start(ctx context.Context) {
	logger := loggerPkg.From(ctx)

	_, err := s.cron.AddFunc("0 * * * *", func() {
		if ctx.Err() != nil {
			logger.Info("Hourly cron skipped: context canceled")
			return
		}
		logger.Info("Hourly cron triggered")
		select {
		case s.events <- "hourly":
		case <-ctx.Done():
			logger.Info("Hourly cron event send canceled")
		}
	})
	if err != nil {
		logger.Errorf("Failed to schedule hourly cron: %v", err)
		return
	}

	_, err = s.cron.AddFunc("0 12 * * *", func() {
		if ctx.Err() != nil {
			logger.Info("Daily cron skipped: context canceled")
			return
		}
		logger.Info("Daily cron triggered")
		select {
		case s.events <- "daily":
		case <-ctx.Done():
			logger.Info("Daily cron event send canceled")
		}
	})
	if err != nil {
		logger.Errorf("Failed to schedule daily cron: %v", err)
		return
	}

	s.cron.Start()

	go func() {
		<-ctx.Done()
		s.Stop(ctx)
	}()
}

func (s *CronEventSource) Events() <-chan string {
	return s.events
}

func (s *CronEventSource) Stop(ctx context.Context) {
	logger := loggerPkg.From(ctx)
	logger.Info("Cron scheduler stopped")
	s.cron.Stop()
	close(s.events)
}
