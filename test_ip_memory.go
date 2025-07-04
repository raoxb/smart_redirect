package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/raoxb/smart_redirect/internal/config"
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
)

func main() {
	// Load config
	cfg, err := config.Load("config/test.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.Database.Postgres.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create services
	linkService := services.NewLinkService(db, redisClient)
	ipMemoryService := services.NewIPMemoryService(redisClient)

	// Create test data
	link := &models.Link{
		LinkID:       "test123",
		BusinessUnit: "bu01",
		Network:      "test",
		IsActive:     true,
		Targets: []models.Target{
			{ID: 1, URL: "https://target1.com", Weight: 30, IsActive: true},
			{ID: 2, URL: "https://target2.com", Weight: 30, IsActive: true},
			{ID: 3, URL: "https://target3.com", Weight: 40, IsActive: true},
		},
	}

	testIP := "192.168.1.100"
	fmt.Printf("Testing IP Memory Service with IP: %s\n\n", testIP)

	// Simulate multiple visits from the same IP
	for i := 1; i <= 5; i++ {
		fmt.Printf("Visit #%d:\n", i)
		
		// Select target using the link service
		selected, err := linkService.SelectTarget(link, testIP, "US")
		if err != nil {
			log.Printf("Error selecting target: %v", err)
			continue
		}
		
		fmt.Printf("  Selected Target ID: %d (URL: %s)\n", selected.ID, selected.URL)
		
		// Show IP stats
		ctx := context.Background()
		stats, err := ipMemoryService.GetIPStats(ctx, testIP)
		if err == nil {
			fmt.Printf("  IP Stats: %+v\n", stats)
		}
		
		fmt.Println()
		time.Sleep(500 * time.Millisecond)
	}

	// Test with a different IP
	testIP2 := "192.168.1.200"
	fmt.Printf("\nTesting with different IP: %s\n", testIP2)
	
	selected, err := linkService.SelectTarget(link, testIP2, "US")
	if err == nil {
		fmt.Printf("  Selected Target ID: %d (URL: %s)\n", selected.ID, selected.URL)
		fmt.Println("  Note: Different IP starts with fresh memory (no visit history)")
	}

	// Clear memory for first IP and test again
	fmt.Printf("\nClearing memory for IP: %s\n", testIP)
	ctx := context.Background()
	err = ipMemoryService.ClearIPMemory(ctx, testIP, link.LinkID)
	if err == nil {
		fmt.Println("  Memory cleared successfully")
		
		selected, err := linkService.SelectTarget(link, testIP, "US")
		if err == nil {
			fmt.Printf("  Next visit after clear - Selected Target ID: %d\n", selected.ID)
			fmt.Println("  Note: Should start fresh as if first visit")
		}
	}

	fmt.Println("\nIP Memory Service Test Complete!")
}