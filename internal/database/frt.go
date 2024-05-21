package database

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Date struct {
	Day   string
	Month string
	Year  string
}

func FetchFRTMaxFetchId() int {
	date := getTodayDate()

	type MaxIdFRTResult struct {
		MaxID int `gorm:"column:max_id"`
	}

	var result MaxIdFRTResult

	queryDate := fmt.Sprintf("%s-%s-01", date.Year, date.Month)
	if date.Day == "01" {
		queryDate = fmt.Sprintf("%s-%s-%s", date.Year, date.Month, date.Day)
	}

	err := TestDB.Raw("select max(cast(frt_log_id as signed)) max_id from frt_logs where log_date > ?", queryDate).Scan(&result).Error
	if err != nil {
		log.Fatalf("failed to fetch max id: %v", err)
	}

	return result.MaxID
}

func getTodayDate() Date {
	today := time.Now()

	return Date{
		Day:   fmt.Sprintf("%02d", today.Day()),
		Month: fmt.Sprintf("%02d", int(today.Month())),
		Year:  fmt.Sprintf("%04d", today.Year()),
	}
}

func InsertFRTLogs(frtData []AwsFRTData) {
	// Create a WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup

	// Load the "Asia/Kolkata" timezone
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	// Create a semaphore with a capacity of 100 to limit concurrency
	sem := make(chan bool, 100)

	// Loop over the frtData slice
	for i := range frtData {
		// Add a count to the WaitGroup
		wg.Add(1)

		// Acquire a semaphore
		sem <- true

		// Launch a goroutine to process the data
		go func(i int) {
			// Ensure the semaphore is released when the function exits
			defer func() { <-sem }()
			// Ensure the WaitGroup count is decremented when the function exits
			defer wg.Done()

			// Adjust the LogDate and FRTCreatedDate to the "Asia/Kolkata" timezone
			_, offset := frtData[i].LogDate.In(loc).Zone()
			frtData[i].LogDate = frtData[i].LogDate.Add(-time.Duration(offset) * time.Second)
			_, offset = frtData[i].FRTCreatedDate.In(loc).Zone()
			frtData[i].FRTCreatedDate = frtData[i].FRTCreatedDate.Add(-time.Duration(offset) * time.Second)

			// Insert the data into the database
			err := TestDB.Exec(`REPLACE INTO frt_logs (device_id, user_id, log_date, log_type, frt_created_date, frt_log_id) VALUES (?, ?, ?, ?, ?, ?)`, frtData[i].DeviceID, frtData[i].UserID, frtData[i].LogDate, frtData[i].LogType, frtData[i].FRTCreatedDate, frtData[i].FRTLogID).Error
			if err != nil {
				log.Fatalf("failed to insert FRT logs: %v", err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Print a success message
	fmt.Println("FRT logs inserted successfully, total records: ", len(frtData))
}
