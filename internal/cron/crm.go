package cron

import (
	"attendance/internal/database"
	"fmt"

	"github.com/robfig/cron"
)

func CrmCron() {
	c := cron.New()

	// Run the CrmPreviousDayCron function every day at 7:30 AM
	c.AddFunc("0 30 7 * * *", func() {
		CrmPreviousDayCron()
	})

	// Run the CrmCurrentDayCron function every day at 9:30 PM
	c.AddFunc("0 30 21 * * *", func() {
		CrmCurrentDayCron()
	})

	c.Start()
}

func CrmPreviousDayCron() {
	attendanceData := database.GetPreviousDayCRMAttendanceData()
	unmatchedData := database.GetPreviousDayUnmatchedCRMAttendanceData()

	// combine both attendanceData and unmatchedData and insert into AWS FRT table
	combinedAttendanceData := append(attendanceData, unmatchedData...)
	if len(combinedAttendanceData) < 1 {
		return
	}
	database.InsertCrmToAwsFrtDataBulk(combinedAttendanceData)
	fmt.Println("CRM Previous Day Cron executed successfully")
}

func CrmCurrentDayCron() {
	attendanceData := database.GetCurrentDayCRMAttendanceData()
	unmatchedData := database.GetCurrentDayUnmatchedCRMAttendanceData()

	// combine both attendanceData and unmatchedData and insert into AWS FRT table
	combinedAttendanceData := append(attendanceData, unmatchedData...)
	if len(combinedAttendanceData) < 1 {
		return
	}
	database.InsertCrmToAwsFrtDataBulk(combinedAttendanceData)
	fmt.Println("CRM Current Day Cron executed successfully")
}
