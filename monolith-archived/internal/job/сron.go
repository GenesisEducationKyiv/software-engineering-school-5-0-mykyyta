package job

import (
	"context"
	"log"

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
	log.Println("[Scheduler] Starting cron scheduler...")

	_, err := s.cron.AddFunc("0 * * * *", func() {
		if ctx.Err() != nil {
			log.Println("[Scheduler] Skipping hourly task due to canceled context")
			return
		}
		log.Println("[Scheduler] Trigger: hourly")
		select {
		case s.events <- "hourly":
		case <-ctx.Done():
			log.Println("[Scheduler] Context canceled while sending hourly event")
		}
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule hourly: %v", err)
		return
	}

	_, err = s.cron.AddFunc("0 12 * * *", func() {
		if ctx.Err() != nil {
			log.Println("[Scheduler] Skipping daily task due to canceled context")
			return
		}
		log.Println("[Scheduler] Trigger: daily")
		select {
		case s.events <- "daily":
		case <-ctx.Done():
			log.Println("[Scheduler] Context canceled while sending daily event")
		}
	})
	if err != nil {
		log.Printf("[Scheduler] Failed to schedule daily: %w", err)
		return
	}

	s.cron.Start()

	go func() {
		<-ctx.Done()
		log.Println("[Scheduler] Context done, stopping scheduler...")
		s.Stop()
	}()
}

func (s *CronEventSource) Events() <-chan string {
	return s.events
}

func (s *CronEventSource) Stop() {
	log.Println("[Scheduler] Stopping cron scheduler...")
	s.cron.Stop()
	close(s.events)
}
