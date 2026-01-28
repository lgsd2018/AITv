package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

type Drama struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string
	Status    string
	CreatedAt time.Time
}

func main() {
	dbPath := "data/drama_generator.db"
	absPath, _ := filepath.Abs(dbPath)
	fmt.Printf("Using database at: %s\n", absPath)

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("Database file does not exist at %s\n", dbPath)
		return
	}

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        dbPath,
	}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	var dramas []Drama
	err = db.Table("dramas").Find(&dramas).Error
	if err != nil {
		fmt.Println("Error querying dramas:", err)
	} else {
		fmt.Println("Found", len(dramas), "dramas:")
		fmt.Println("------------------------------------------------")
		for _, d := range dramas {
			fmt.Printf("ID: %d\nTitle: %s\nStatus: %s\nCreated: %s\n",
				d.ID, d.Title, d.Status, d.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println("------------------------------------------------")
		}
	}
}
