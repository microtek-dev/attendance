package main

import (
	"attendance/internal/database"
	"attendance/internal/server"
	"fmt"
)

func main() {
	server := server.NewServer()

	database.SyncSalesAttendanceFromFieldAssist()
	// cron.FRTCron()
	// cron.PeoplestrongCron()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
