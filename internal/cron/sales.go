package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func StartCron() {
	c := cron.New()
	// run this every 1 minute
	c.AddFunc("@every 1m", func() {
		database.SyncEmployeeData()
	})
	// c.Start()
}
