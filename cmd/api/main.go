package main

import (
	"attendance/internal/cron"
	"attendance/internal/database"
	"attendance/internal/server"
	"fmt"
)

func main() {
	server := server.NewServer()

	cron.StartCron()
	database.SyncEmployeeData()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
