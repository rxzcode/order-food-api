package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"order-food-api/core/config"
)

func Connect(dbCfg config.DBConfig) *gorm.DB {
	if dbCfg.Host == "db" && os.Getenv("RUNNING_IN_DOCKER") != "true" {
		dbCfg.Host = "localhost"
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name)

	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Retrying DB connection... (%d/5)", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}
