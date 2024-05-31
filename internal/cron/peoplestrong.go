package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func PeoplestrongCron() {
	c := cron.New()

	// Run the sync every 10 minutes
	c.AddFunc("0 */12 * * * *", func() {
		SyncFrtLogsPeoplestrongCron()
	})

	c.Start()
}

func SyncFrtLogsPeoplestrongCron() {
	lastPushedId := database.LastPushedId()
	frtLogData := database.FetchFrtData(lastPushedId)

	// if there is no data to push, return
	if len(frtLogData) == 0 {
		return
	}

	currentPushedId := database.InsertFrtLogBulk(frtLogData)
	database.UpdateLastPushedId(currentPushedId)
}
