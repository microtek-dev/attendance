package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func StartCron() {
	c := cron.New()

	// Sync employee data every day at 5:00 AM
	c.AddFunc("0 5 * * *", func() {
		database.SyncEmployeeData()
	})
	// c.Start()
}
