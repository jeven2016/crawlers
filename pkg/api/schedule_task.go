package api

import (
	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	"time"
)

type ScheduleTask struct {
}

func (t *ScheduleTask) Run() (err error) {
	schd := gocron.NewScheduler(time.Local)
	_, err = schd.Every(10).Minute().Do(func() {
		zap.S().Info("Scheduler runs")
	})
	if err != nil {
		return err
	}
	schd.StartAsync()
	return
}
