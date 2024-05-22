package models

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB // グローバル変数としてデータベース接続を保持

func SetupDB() (*gorm.DB, error) {
	dbUser := "postgres"
	dbPassword := "postgres"
	dbName := "postgres"
	dbHost := "db"
	dbPort := "5432"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo", dbHost, dbUser, dbPassword, dbName, dbPort)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.AutoMigrate(&StockFrames{}, &Parts{}, &Locations{}, &PartLocations{}, &Orders{}, &User{})
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}
	return db, nil
}

func GetDB() *gorm.DB {
	return db
}
