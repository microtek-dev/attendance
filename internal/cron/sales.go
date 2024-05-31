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
		database.SyncSalesAttendanceFromFieldAssist()
	})

	// Sync sales attendance every day at 6:00 AM
	c.AddFunc("00 6 * * *", func() {
		database.SyncSalesAttendanceFromFieldAssist()
	})

	// Sync sales attendance every day at 6:03 AM
	c.AddFunc("03 6 * * *", func() {
		syncSalesAttendanceWithFrtYesterday()
	})

	// Sync sales attendance every day at 9:30 AM
	c.AddFunc("30 21 * * *", func() {
		syncSalesAttendanceWithFrtToday()
	})

	// Sync sales attendance every day at 8:59 PM
	c.AddFunc("59 20 * * *", func() {
		syncSalesAttendanceWithFrtToday()
	})

	// c.Start()
}

func syncSalesAttendanceWithFrtYesterday() {
	salesAttendanceData := database.GetSalesAttendanceFromDailyTask("yesterday")
	database.SaveSalesAttendanceLocallyBulk(salesAttendanceData)
	database.InsertSalesToAwsFrtDataBulk(salesAttendanceData)

	// process unmatched sales attendance
	unmatchedSalesAttendance := database.GetSalesAttendanceFromDailyTaskUnmatched("yesterday")
	database.SaveSalesAttendanceLocallyBulk(unmatchedSalesAttendance)
	database.InsertSalesToAwsFrtDataBulk(unmatchedSalesAttendance)
}

func syncSalesAttendanceWithFrtToday() {
	salesAttendanceData := database.GetSalesAttendanceFromDailyTask("today")
	database.SaveSalesAttendanceLocallyBulk(salesAttendanceData)
	database.InsertSalesToAwsFrtDataBulk(salesAttendanceData)

	// process unmatched sales attendance
	unmatchedSalesAttendance := database.GetSalesAttendanceFromDailyTaskUnmatched("today")
	database.SaveSalesAttendanceLocallyBulk(unmatchedSalesAttendance)
	database.InsertSalesToAwsFrtDataBulk(unmatchedSalesAttendance)
}
