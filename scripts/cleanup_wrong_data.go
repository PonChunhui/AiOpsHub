package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "host=192.168.100.10 port=5432 user=aiops password=aiops123 dbname=aiopsdb sslmode=disable TimeZone=Asia/Shanghai"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	defer sqlDB.Close()

	result := db.Exec("DELETE FROM knowledges WHERE created_by = 'system'")
	if result.Error != nil {
		log.Printf("Failed to delete: %v", result.Error)
	} else {
		log.Printf("Deleted %d rows from knowledges table", result.RowsAffected)
	}

	var count int64
	db.Table("knowledges").Count(&count)
	fmt.Printf("Remaining knowledges count: %d\n", count)
}
