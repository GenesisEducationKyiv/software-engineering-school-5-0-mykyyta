package jobs

import (
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

func (s *CronEventSource) Start() {
	log.Println("[Scheduler] Starting cron scheduler...")

	_, err := s.cron.AddFunc("0 * * * *", func() {
		log.Println("[Scheduler] Trigger: hourly")
		s.events <- "hourly"
	})
	if err != nil {
		log.Fatalf("[Scheduler] Failed to schedule hourly: %v", err)
	}

	_, err = s.cron.AddFunc("0 12 * * *", func() {
		log.Println("[Scheduler] Trigger: daily")
		s.events <- "daily"
	})
	if err != nil {
		log.Fatalf("[Scheduler] Failed to schedule daily: %v", err)
	}

	s.cron.Start()
}

func (s *CronEventSource) Events() <-chan string {
	return s.events
}

func (s *CronEventSource) Stop() {
	log.Println("[Scheduler] Stopping cron scheduler...")
	s.cron.Stop()
	close(s.events)
}
