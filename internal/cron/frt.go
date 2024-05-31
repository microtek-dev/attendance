package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func FRTCron() {
	c := cron.New()

	// Run the sync every 5 minutes
	c.AddFunc("0 */5 * * * *", func() {
		SyncAwsFrtDataCron()
	})

	c.Start()
}

func SyncAwsFrtDataCron() {
	maxId := database.FetchFRTMaxFetchId()
	frtData := database.FetchAwsFRTData(maxId)

	// if no data is fetched, return
	if len(frtData) == 0 {
		return
	}

	database.InsertFRTLogs(frtData)
}
