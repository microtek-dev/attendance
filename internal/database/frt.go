package database

import (
	"fmt"
	"log"
	"strconv"
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

type FRTData struct {
	FRTLogID       int       `gorm:"column:frt_log_id"`
	DeviceID       int       `gorm:"column:device_id"`
	UserID         string    `gorm:"column:user_id"`
	LogDate        time.Time `gorm:"column:log_date"`
	LogType        string    `gorm:"column:log_type"`
	FRTCreatedDate time.Time `gorm:"column:frt_created_date"`
}

func FetchFRTData(maxFetchID int) []FRTData {
	// Get today's date
	date := getTodayDate()

	// Extract the month and year
	month := date.Month
	year := date.Year

	// Initialize a slice to hold the fetched data
	var frtData []FRTData

	// Convert the month and year to integers
	monthInt, err := strconv.Atoi(month)
	if err != nil {
		log.Fatalf("failed to convert month to integer: %v", err)
	}

	yearInt, err := strconv.Atoi(year)
	if err != nil {
		log.Fatalf("failed to convert year to integer: %v", err)
	}

	// Construct the table name based on the current month and year
	tableName := fmt.Sprintf("DeviceLogs_%d_%d", monthInt, yearInt)

	// Construct the SQL query
	query := fmt.Sprintf(`SELECT TOP 10000 DeviceLogId frt_log_id, DeviceId device_id, UserId user_id, LogDate log_date, C1 log_type, CreatedDate frt_created_date FROM %s WHERE DeviceLogId > ? ORDER BY DeviceLogId`, tableName)

	// Execute the query and scan the results into the frtData slice
	err = AwsDB.Raw(query, maxFetchID).Scan(&frtData).Error
	if err != nil {
		log.Fatalf("failed to fetch FRT data: %v", err)
	}

	// Print the number of records fetched
	fmt.Println("Total records fetched: ", len(frtData))

	// Return the fetched data
	return frtData
}

func InsertFRTLogs(frtData []FRTData) {
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
