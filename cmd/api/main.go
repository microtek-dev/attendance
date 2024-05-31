package main

import (
	"attendance/internal/cron"
	"attendance/internal/server"
	"fmt"
)

func main() {
	server := server.NewServer()

	cron.FRTCron()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
