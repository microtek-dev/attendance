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
	lastPushedId := database.LastPushedId()
	fmt.Println("Last pushed id: ", lastPushedId)

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
