package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/raoxb/smart_redirect/internal/config"
)

// Simple models for testing
type Link struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	LinkID       string `gorm:"uniqueIndex;size:10" json:"link_id"`
	BusinessUnit string `gorm:"size:10" json:"business_unit"`
	Network      string `gorm:"size:50" json:"network"`
	TotalCap     int    `json:"total_cap"`
	CurrentHits  int    `json:"current_hits"`
	BackupURL    string `json:"backup_url"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`
}

type Target struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	LinkID       uint   `gorm:"index" json:"link_id"`
	URL          string `json:"url"`
	Weight       int    `json:"weight"`
	Cap          int    `json:"cap"`
	CurrentHits  int    `json:"current_hits"`
	Countries    string `json:"countries"`
	ParamMapping string `gorm:"type:jsonb" json:"param_mapping"`
	StaticParams string `gorm:"type:jsonb" json:"static_params"`
	IsActive     bool   `gorm:"default:true" json:"is_active"`
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	
	// Load config
	cfg, err := config.Load("config/test.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database directly
	db, err := gorm.Open(postgres.Open(cfg.Database.Postgres.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"timestamp": time.Now(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	
	// Mock auth endpoint
	v1.POST("/auth/login", func(c *gin.Context) {
		var request struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		
		if request.Username == "admin" && request.Password == "admin123" {
			c.JSON(200, gin.H{
				"token": "mock-jwt-token",
				"user": gin.H{
					"id": 1,
					"username": "admin",
					"role": "admin",
				},
			})
		} else {
			c.JSON(401, gin.H{"error": "Invalid credentials"})
		}
	})

	// Links endpoint
	v1.GET("/links", func(c *gin.Context) {
		var links []Link
		db.Find(&links)
		c.JSON(200, gin.H{"data": links})
	})

	// Get specific link with targets
	v1.GET("/links/:id", func(c *gin.Context) {
		linkID := c.Param("id")
		
		var link Link
		if err := db.Where("link_id = ?", linkID).First(&link).Error; err != nil {
			c.JSON(404, gin.H{"error": "Link not found"})
			return
		}
		
		var targets []Target
		db.Where("link_id = ?", link.ID).Find(&targets)
		
		c.JSON(200, gin.H{
			"link": link,
			"targets": targets,
		})
	})

	// Test redirect endpoint  
	v1.GET("/redirect/:bu/:link_id", func(c *gin.Context) {
		bu := c.Param("bu")
		linkID := c.Param("link_id")
		network := c.Query("network")
		
		var link Link
		if err := db.Where("business_unit = ? AND link_id = ?", bu, linkID).First(&link).Error; err != nil {
			c.JSON(404, gin.H{"error": "Link not found"})
			return
		}
		
		var targets []Target
		db.Where("link_id = ? AND is_active = ?", link.ID, true).Find(&targets)
		
		if len(targets) == 0 {
			if link.BackupURL != "" {
				c.Redirect(http.StatusFound, link.BackupURL)
				return
			}
			c.JSON(404, gin.H{"error": "No active targets"})
			return
		}
		
		// Simple target selection (first active target)
		target := targets[0]
		
		// Parse and apply parameter mapping
		var paramMap map[string]string
		if target.ParamMapping != "" {
			json.Unmarshal([]byte(target.ParamMapping), &paramMap)
		}
		
		var staticParams map[string]string
		if target.StaticParams != "" {
			json.Unmarshal([]byte(target.StaticParams), &staticParams)
		}
		
		// Build redirect URL with parameters
		redirectURL := target.URL
		queryParams := "?"
		
		// Add query parameters from request
		for key, values := range c.Request.URL.Query() {
			if mappedKey, exists := paramMap[key]; exists {
				queryParams += fmt.Sprintf("%s=%s&", mappedKey, values[0])
			} else {
				queryParams += fmt.Sprintf("%s=%s&", key, values[0])
			}
		}
		
		// Add static parameters
		for key, value := range staticParams {
			queryParams += fmt.Sprintf("%s=%s&", key, value)
		}
		
		// Add network parameter
		if network != "" {
			queryParams += fmt.Sprintf("network=%s&", network)
		}
		
		if len(queryParams) > 1 {
			redirectURL += queryParams[:len(queryParams)-1] // Remove trailing &
		}
		
		// Update hit count
		db.Model(&link).Update("current_hits", link.CurrentHits+1)
		db.Model(&target).Update("current_hits", target.CurrentHits+1)
		
		c.Header("X-Redirect-Info", fmt.Sprintf("Link: %s, Target: %d", linkID, target.ID))
		c.Redirect(http.StatusFound, redirectURL)
	})

	// System stats endpoint
	v1.GET("/stats/system", func(c *gin.Context) {
		var linkCount int64
		var targetCount int64
		
		db.Model(&Link{}).Count(&linkCount)
		db.Model(&Target{}).Count(&targetCount)
		
		c.JSON(200, gin.H{
			"total_links": linkCount,
			"total_targets": targetCount,
			"active_links": linkCount, // Simplified
			"total_redirects": 42,     // Mock data
		})
	})

	log.Println("Test server starting on :8080")
	log.Println("Available endpoints:")
	log.Println("  GET  /health")
	log.Println("  POST /api/v1/auth/login")
	log.Println("  GET  /api/v1/links")
	log.Println("  GET  /api/v1/links/:id") 
	log.Println("  GET  /api/v1/redirect/:bu/:link_id?network=...")
	log.Println("  GET  /api/v1/stats/system")
	
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}