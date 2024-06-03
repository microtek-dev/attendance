package cron

import (
	"attendance/internal/database"
	"fmt"
	"time"

	"github.com/robfig/cron"
)

func SalesCron() {
	c := cron.New()

	// Sync employee data every day at 5:00 AM
	c.AddFunc("0 0 5 * * *", func() {
		database.SyncEmployeeData()
	})

	// Sync sales attendance every day at 5:30 AM
	c.AddFunc("0 30 5 * * *", func() {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		database.SyncSalesAttendanceFromFieldAssist(yesterday)
	})

	// Sync sales attendance every day at 6:00 AM
	c.AddFunc("0 0 6 * * *", func() {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		database.SyncSalesAttendanceFromFieldAssist(yesterday)
	})

	// Sync today's sales attendance every day at 8:00 AM
	c.AddFunc("0 0 8 * * *", func() {
		today := time.Now().Format("2006-01-02")
		database.SyncSalesAttendanceFromFieldAssist(today)
	})

	// Sync today's sales attendance every day at 8:15 PM
	c.AddFunc("0 15 20 * * *", func() {
		today := time.Now().Format("2006-01-02")
		database.SyncSalesAttendanceFromFieldAssist(today)
	})

	// Sync today's sales attendance every day at 8:30 PM
	c.AddFunc("0 30 20 * * *", func() {
		today := time.Now().Format("2006-01-02")
		database.SyncSalesAttendanceFromFieldAssist(today)
	})

	// Sync sales attendance every day at 6:03 AM
	c.AddFunc("0 3 6 * * *", func() {
		syncSalesAttendanceWithFrtYesterday()
	})

	// Sync sales attendance every day at 9:30 AM
	c.AddFunc("0 30 9 * * *", func() {
		SyncSalesAttendanceWithFrtToday()
	})

	// Sync sales attendance every day at 8:59 PM
	c.AddFunc("0 59 20 * * *", func() {
		SyncSalesAttendanceWithFrtToday()
	})

	c.Start()
}

func syncSalesAttendanceWithFrtYesterday() {
	fmt.Println(time.Now(), "Syncing sales attendance with FRT for yesterday")
	salesAttendanceData := database.GetSalesAttendanceFromDailyTask("yesterday")
	database.SaveSalesAttendanceLocallyBulk(salesAttendanceData)
	database.InsertSalesToAwsFrtDataBulk(salesAttendanceData)

	// process unmatched sales attendance
	unmatchedSalesAttendance := database.GetSalesAttendanceFromDailyTaskUnmatched("yesterday")
	database.SaveSalesAttendanceLocallyBulk(unmatchedSalesAttendance)
	database.InsertSalesToAwsFrtDataBulk(unmatchedSalesAttendance)
	fmt.Println(time.Now(), "Syncing sales attendance with FRT for yesterday done")
}

func SyncSalesAttendanceWithFrtToday() {
	fmt.Println(time.Now(), "Syncing sales attendance with FRT for today")
	salesAttendanceData := database.GetSalesAttendanceFromDailyTask("today")
	database.SaveSalesAttendanceLocallyBulk(salesAttendanceData)
	database.InsertSalesToAwsFrtDataBulk(salesAttendanceData)

	// process unmatched sales attendance
	unmatchedSalesAttendance := database.GetSalesAttendanceFromDailyTaskUnmatched("today")
	database.SaveSalesAttendanceLocallyBulk(unmatchedSalesAttendance)
	database.InsertSalesToAwsFrtDataBulk(unmatchedSalesAttendance)
	fmt.Println(time.Now(), "Syncing sales attendance with FRT for today done")
}
