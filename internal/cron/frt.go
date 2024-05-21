package cron

import "attendance/internal/database"

func SyncAwsFrtDataCron() {
	maxId := database.FetchFRTMaxFetchId()
	frtData := database.FetchAwsFRTData(maxId)
	database.InsertFRTLogs(frtData)
}
