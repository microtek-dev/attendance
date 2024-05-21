package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func SalesCron() {
	c := cron.New()

	// Sync employee data every day at 5:00 AM
	c.AddFunc("0 5 * * *", func() {
		database.SyncEmployeeData()
	})

	// Sync sales attendance every day at 5:30 AM
	c.AddFunc("31 5 * * *", func() {
		database.SyncSalesAttendance()
	})

	// Sync sales attendance every day at 6:00 AM
	c.AddFunc("00 6 * * *", func() {
		database.SyncSalesAttendance()
	})

	// c.Start()
}
