package database

import (
	"fmt"
	"log"
	"sync"
	"time"
)

func LastPushedId() int {
	type LastPushedIdResult struct {
		LastFetchId int `gorm:"column:param1_int"`
	}

	var result LastPushedIdResult

	err := ProgressionDB.Raw(`select param1_int from mtek_db.app_config_master where config_name = "FRT_TO_PS"`).Scan(&result).Error
	if err != nil {
		log.Fatalf("failed to fetch max id: %v", err)
	}

	return result.LastFetchId
}

type FRTLogData struct {
	FrtID   int       `gorm:"column:log_id"`
	UserID  string    `gorm:"column:user_id"`
	LogDate time.Time `gorm:"column:log_date"`
}

func FetchFrtData(lastPushedId int) []FRTLogData {
	var frtLogData []FRTLogData

	err := ProgressionDB.Raw(`select log_id, user_id, log_date from mtek_db.frt_logs where log_id > ? and (user_id like '5%' or user_id like '6%' or user_id like '7%') order by log_id asc`, lastPushedId).Scan(&frtLogData).Error
	if err != nil {
		log.Fatalf("failed to fetch frt log data: %v", err)
	}

	return frtLogData
}

func InsertPunchData(empID string, punchDateTime time.Time) {
	err := AwsDB.Exec(`INSERT INTO AttendanceToPS(Empid ,Punch_Date_Time ,RP_CREATED_DATE ,Record_LastUpdated ,Isread) VALUES (?, ?, GETDATE(), GETDATE(), '0')`, empID, punchDateTime).Error
	if err != nil {
		log.Fatalf("failed to insert punch data: %v", err)
	}
}

func InsertFrtLogBulk(punchData []FRTLogData) int {
	var wg sync.WaitGroup
	chunkSize := 100 // Adjust this value based on your needs

	// Split punchData into smaller slices
	for i := 0; i < len(punchData); i += chunkSize {
		end := i + chunkSize

		if end > len(punchData) {
			end = len(punchData)
		}

		wg.Add(1)
		go func(chunk []FRTLogData) {
			defer wg.Done()

			for _, data := range chunk {
				InsertPunchData(data.UserID, data.LogDate)
			}
		}(punchData[i:end])
	}

	wg.Wait()

	fmt.Println("Total records inserted to AttendanceToPS table: ", len(punchData))

	return punchData[len(punchData)-1].FrtID
}

func UpdateLastPushedId(lastPushedId int) {
	err := ProgressionDB.Exec(`update mtek_db.app_config_master set param1_int = ?, updated_at = CURRENT_TIMESTAMP() where config_name = "FRT_TO_PS"`, lastPushedId).Error
	if err != nil {
		log.Fatalf("failed to update last pushed id: %v", err)
	}
}

// push attendance from device_log
