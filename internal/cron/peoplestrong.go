package cron

import (
	"attendance/internal/database"

	"github.com/robfig/cron"
)

func PeoplestrongCron() {
	c := cron.New()

	// Run the sync every 12 minutes
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
	// some of the UserID in the frtLogData have a trailing "T" character, only take the first five characters, if more than 5 characters
	for i := 0; i < len(frtLogData); i++ {
		if len(frtLogData[i].UserID) > 5 {
			frtLogData[i].UserID = frtLogData[i].UserID[:5]
		}
	}

	currentPushedId := database.InsertFrtLogBulk(frtLogData)
	database.UpdateLastPushedId(currentPushedId)
}
