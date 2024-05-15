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
	maxId := database.FetchFRTMaxFetchId()
	database.FetchFRTData(maxId)

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
