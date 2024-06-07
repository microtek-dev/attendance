package main

import (
	"attendance/internal/database"
	"attendance/internal/server"
	"fmt"
	"time"
)

func main() {
	server := server.NewServer()

	// cron.FRTCron()
	// cron.PeoplestrongCron()
	// cron.SalesCron()

	// set date as 29 May 2024 in string format
	date := time.Date(2024, time.May, 29, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	database.SyncSalesAttendanceFromFieldAssist(date)

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
