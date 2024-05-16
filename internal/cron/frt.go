package cron

import "attendance/internal/database"

func SyncAwsFrtDataCron() {
	maxId := database.FetchFRTMaxFetchId()
	frtData := database.FetchFRTData(maxId)
	database.InsertFRTLogs(frtData)
}
