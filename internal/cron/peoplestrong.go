package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func PeoplestrongCron() {
	c := cron.New()

	// Run the sync every 10 minutes
	c.AddFunc("0 */10 * * * *", func() {
		SyncFrtLogsPeoplestrongCron()
	})

	c.Start()
}

func SyncFrtLogsPeoplestrongCron() {
	lastPushedId := database.LastPushedId()
	frtLogData := database.FetchFrtData(lastPushedId)
	currentPushedId := database.InsertFrtLogBulk(frtLogData)
	database.UpdateLastPushedId(currentPushedId)
}
