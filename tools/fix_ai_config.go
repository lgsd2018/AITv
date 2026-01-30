//go:build tools
// +build tools

package main

import (
	"fmt"
	"log"
	"path/filepath"
	"os"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/infrastructure/database"
	"github.com/drama-generator/backend/pkg/config"
)

func main() {
	fmt.Println("Starting AI Config Fix...")

	cwd, _ := os.Getwd()
	fmt.Println("Current Working Directory:", cwd)
	
	dbPath := filepath.Join(cwd, "data", "drama_generator.db")
	fmt.Println("Database Path:", dbPath)

	// 1. Setup DB Config
	dbConfig := config.DatabaseConfig{
		Type: "sqlite",
		Path: dbPath,
	}

	// 2. Connect to Database
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	fmt.Println("Database connected successfully.")

	// 3. Auto Migrate (Ensure tables exist)
	fmt.Println("Running AutoMigrate...")
	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	fmt.Println("Migration completed.")

	// 4. Find configs for video generation with provider volces/doubao
	var configs []models.AIServiceConfig
	err = db.Where("service_type = ? AND provider IN ?", "video", []string{"doubao", "volcengine", "volces"}).Find(&configs).Error
	if err != nil {
		log.Fatalf("Failed to query configs: %v", err)
	}

	// Correct values
	correctKey := "c3f7f4b0-5072-47ee-b944-0a73f2377443"
	correctBaseURL := "https://ark.cn-beijing.volces.com"
	correctEndpoint := "/api/v3/contents/generations/tasks"
	correctQueryEndpoint := "/api/v3/contents/generations/tasks/{taskId}"
	targetModel := "doubao-seedance-1-5-pro-251215"

	if len(configs) == 0 {
		fmt.Println("No Volces/Doubao video config found. Creating a new one...")
		newConfig := models.AIServiceConfig{
			ServiceType:   "video",
			Name:          "VolcEngine Video",
			Provider:      "volces",
			BaseURL:       correctBaseURL,
			APIKey:        correctKey,
			Endpoint:      correctEndpoint,
			QueryEndpoint: correctQueryEndpoint,
			Model:         models.ModelField{targetModel},
			IsActive:      true,
			IsDefault:     true,
			Priority:      100,
		}
		if err := db.Create(&newConfig).Error; err != nil {
			log.Fatalf("Failed to create new config: %v", err)
		}
		fmt.Println("Created new config successfully.")
	} else {
		fmt.Printf("Found %d configs. Updating...\n", len(configs))
		for _, c := range configs {
			fmt.Printf("Checking config ID: %d, Name: %s\n", c.ID, c.Name)
			
			needsUpdate := false
			updates := map[string]interface{}{}

			if c.APIKey != correctKey {
				updates["api_key"] = correctKey
				needsUpdate = true
				fmt.Println("  - Updating API Key")
			}
			if c.BaseURL != correctBaseURL {
				updates["base_url"] = correctBaseURL
				needsUpdate = true
				fmt.Println("  - Updating Base URL")
			}
			if c.Endpoint != correctEndpoint {
				updates["endpoint"] = correctEndpoint
				needsUpdate = true
				fmt.Println("  - Updating Endpoint")
			}
			if c.QueryEndpoint != correctQueryEndpoint {
				updates["query_endpoint"] = correctQueryEndpoint
				needsUpdate = true
				fmt.Println("  - Updating Query Endpoint")
			}

			// Check model
			hasModel := false
			for _, m := range c.Model {
				if m == targetModel {
					hasModel = true
					break
				}
			}
			if !hasModel {
				newModels := append(c.Model, targetModel)
				updates["model"] = newModels
				needsUpdate = true
				fmt.Println("  - Adding Model: " + targetModel)
			}

			if needsUpdate {
				if err := db.Model(&c).Updates(updates).Error; err != nil {
					log.Printf("Failed to update config ID %d: %v\n", c.ID, err)
				} else {
					fmt.Printf("Successfully updated config ID %d\n", c.ID)
				}
			} else {
				fmt.Println("  - Config is already correct.")
			}
		}
	}

	fmt.Println("Done.")
}
