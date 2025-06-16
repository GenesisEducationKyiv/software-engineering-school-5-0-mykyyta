package jobs

import (
	"log"

	"github.com/robfig/cron/v3"
)

type CallbackFunc func(frequency string)

type CronScheduler struct {
	callback CallbackFunc
	cron     *cron.Cron
}

func NewCronScheduler(cb CallbackFunc) *CronScheduler {
	return &CronScheduler{
		callback: cb,
		cron:     cron.New(),
	}
}

func (s *CronScheduler) Start() {
	log.Println("[Scheduler] Starting cron scheduler...")

	_, err := s.cron.AddFunc("0 * * * *", func() {
		log.Println("[Scheduler] Trigger: hourly")
		s.callback("hourly")
	})
	if err != nil {
		log.Fatalf("[Scheduler] Failed to schedule hourly: %v", err)
	}

	_, err = s.cron.AddFunc("0 12 * * *", func() {
		log.Println("[Scheduler] Trigger: daily")
		s.callback("daily")
	})
	if err != nil {
		log.Fatalf("[Scheduler] Failed to schedule daily: %v", err)
	}

	s.cron.Start()
}

func (s *CronScheduler) Stop() {
	log.Println("[Scheduler] Stopping cron scheduler...")
	s.cron.Stop()
}
