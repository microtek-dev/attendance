package database

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

type AwsFRTData struct {
	FRTLogID       int       `gorm:"column:frt_log_id"`
	DeviceID       int       `gorm:"column:device_id"`
	UserID         string    `gorm:"column:user_id"`
	LogDate        time.Time `gorm:"column:log_date"`
	LogType        string    `gorm:"column:log_type"`
	FRTCreatedDate time.Time `gorm:"column:frt_created_date"`
}

func AwsTableName() string {
	// Get today's date
	date := getTodayDate()

	// Extract the month and year
	month := date.Month
	year := date.Year

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

	return tableName
}

func InsertIntoAwsFrtData(userId string, logDate time.Time) {
	// Get the table name
	tableName := AwsTableName()

	// Insert the data into the AWS table
	err := AwsDB.Exec(`INSERT INTO `+tableName+`(UserId, LogDate, CreatedDate) VALUES (?, ?, ?)`, userId, logDate, time.Now()).Error
	if err != nil {
		log.Fatalf("failed to insert into AWS frt data: %v", err)
	}

	log.Printf("Inserted data into AWS: UserId: %s, LogDate: %s", userId, logDate)
}

func FetchAwsFRTData(maxFetchID int) []AwsFRTData {
	// Initialize a slice to hold the fetched data
	var frtData []AwsFRTData

	tableName := AwsTableName()

	// Construct the SQL query
	query := fmt.Sprintf(`SELECT TOP 10000 DeviceLogId frt_log_id, DeviceId device_id, UserId user_id, LogDate log_date, C1 log_type, CreatedDate frt_created_date FROM %s WHERE DeviceLogId > ? ORDER BY DeviceLogId`, tableName)

	// Execute the query and scan the results into the frtData slice
	err := AwsDB.Raw(query, maxFetchID).Scan(&frtData).Error
	if err != nil {
		log.Fatalf("failed to fetch FRT data: %v", err)
	}

	// Print the number of records fetched
	fmt.Println("Total records fetched: ", len(frtData))

	// Return the fetched data
	return frtData
}
