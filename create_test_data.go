package main

import (
	"fmt"
	"log"
	
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Link struct {
	ID           uint   `gorm:"primaryKey"`
	LinkID       string `gorm:"uniqueIndex"`
	BusinessUnit string
	Network      string
	TotalCap     int
	CurrentHits  int
	BackupURL    string
	IsActive     bool
}

type Target struct {
	ID           uint   `gorm:"primaryKey"`
	LinkID       uint
	URL          string
	Weight       int
	Cap          int
	CurrentHits  int
	Countries    string
	ParamMapping string
	StaticParams string
	IsActive     bool
}

func main() {
	dsn := "host=localhost user=postgres password=password dbname=smart_redirect port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create test link
	link := Link{
		LinkID:       generateLinkID(),
		BusinessUnit: "bu01",
		Network:      "google",
		TotalCap:     10000,
		CurrentHits:  0,
		BackupURL:    "https://www.google.com",
		IsActive:     true,
	}
	
	if err := db.Create(&link).Error; err != nil {
		log.Printf("Failed to create link: %v", err)
		return
	}
	
	fmt.Printf("Created link: %s\n", link.LinkID)
	
	// Create targets
	targets := []Target{
		{
			LinkID:       link.ID,
			URL:          "https://example.com/landing1",
			Weight:       40,
			Cap:          5000,
			CurrentHits:  0,
			Countries:    `["US", "CA", "UK"]`,
			ParamMapping: `{}`,
			StaticParams: `{"utm_source":"redirect","utm_medium":"link"}`,
			IsActive:     true,
		},
		{
			LinkID:       link.ID,
			URL:          "https://example.com/landing2",
			Weight:       30,
			Cap:          3000,
			CurrentHits:  0,
			Countries:    `["DE", "FR", "IT"]`,
			ParamMapping: `{}`,
			StaticParams: `{"utm_source":"redirect","utm_campaign":"europe"}`,
			IsActive:     true,
		},
		{
			LinkID:       link.ID,
			URL:          "https://example.com/landing3",
			Weight:       30,
			Cap:          0, // unlimited
			CurrentHits:  0,
			Countries:    `[]`, // all countries
			ParamMapping: `{"src":"source","cmp":"campaign"}`,
			StaticParams: `{"ref":"smart_redirect"}`,
			IsActive:     true,
		},
	}
	
	for i, target := range targets {
		if err := db.Create(&target).Error; err != nil {
			log.Printf("Failed to create target %d: %v", i+1, err)
		} else {
			fmt.Printf("Created target %d: %s\n", i+1, target.URL)
		}
	}
	
	// Print test URL
	fmt.Println("\n========================================")
	fmt.Printf("Test URL: http://103.14.79.22:8080/v1/%s/%s?src=test&cmp=demo\n", link.BusinessUnit, link.LinkID)
	fmt.Println("========================================")
	fmt.Println("\nAdmin Panel: http://103.14.79.22:3003/")
	fmt.Println("Username: admin")
	fmt.Println("Password: admin123")
}

func generateLinkID() string {
	id := uuid.New().String()
	return id[:6]
}