package cron

import "attendance/internal/database"

func SyncFrtLogsPeoplestrongCron() {
	lastPushedId := database.LastPushedId()
	frtLogData := database.FetchFrtData(lastPushedId)
	currentPushedId := database.InsertFrtLogBulk(frtLogData)
	database.UpdateLastPushedId(currentPushedId)
}
