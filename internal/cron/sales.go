package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func StartCron() {
	c := cron.New()
	c.AddFunc("@every 30s", func() {
		database.SyncEmployeeData()
	})
	// c.Start()
}
