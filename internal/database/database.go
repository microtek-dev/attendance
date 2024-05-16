package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"

	_ "github.com/joho/godotenv/autoload"
)

var (
	TestDB         *gorm.DB
	ProgressionDB  *gorm.DB
	AwsDB          *gorm.DB
	PeoplestrongDB *gorm.DB
)

func init() {
	var err error

	TestDB, err = gorm.Open(mysql.Open(os.Getenv("TEST_DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to Test database: %v", err)
	}
	ProgressionDB, err = gorm.Open(mysql.Open(os.Getenv("PROGRESSION_DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to Progression database: %v", err)
	}
	AwsDB, err = gorm.Open(sqlserver.Open(os.Getenv("AWS_DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to AwsDB database: %v", err)
	}
	PeoplestrongDB, err = gorm.Open(sqlserver.Open(os.Getenv("PEOPLESTRONG_DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to Peoplestrong database: %v", err)
	}

	dbs := []*gorm.DB{TestDB, ProgressionDB, AwsDB, PeoplestrongDB}

	for _, db := range dbs {
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("failed to get db instance: %v", err)
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		if err := sqlDB.Ping(); err != nil {
			log.Fatalf("failed to ping database: %v", err)
		}
	}
}
