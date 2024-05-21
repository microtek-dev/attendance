package database

import (
	"fmt"
	"log"
	"strconv"
)

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

func InsertIntoAwsFrtData() {}
