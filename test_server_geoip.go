package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/raoxb/smart_redirect/internal/config"
	"github.com/raoxb/smart_redirect/pkg/geoip"
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

	// Create GeoIP provider
	geoProvider, err := geoip.NewProvider(&cfg.GeoIP)
	if err != nil {
		log.Fatal("Failed to create GeoIP provider:", err)
	}
	defer geoProvider.Close()

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.Database.Postgres.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Setup Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// GeoIP test endpoint
	router.GET("/geoip/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		location, err := geoProvider.GetLocation(ip)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, location)
	})

	// Test redirect endpoint with GeoIP filtering
	v1 := router.Group("/api/v1")
	v1.GET("/redirect/:bu/:link_id", func(c *gin.Context) {
		bu := c.Param("bu")
		linkID := c.Param("link_id")
		network := c.Query("network")
		
		// Get client IP
		clientIP := c.ClientIP()
		if forwardedIP := c.GetHeader("X-Forwarded-For"); forwardedIP != "" {
			clientIP = strings.Split(forwardedIP, ",")[0]
		}
		
		// Get location
		location, err := geoProvider.GetLocation(clientIP)
		if err != nil {
			log.Printf("Failed to get location for IP %s: %v", clientIP, err)
			location = &geoip.Location{
				IP:          clientIP,
				CountryCode: "XX",
			}
		}
		
		var link Link
		if err := db.Where("business_unit = ? AND link_id = ?", bu, linkID).First(&link).Error; err != nil {
			c.JSON(404, gin.H{"error": "Link not found"})
			return
		}
		
		// Get active targets
		var targets []Target
		db.Where("link_id = ? AND is_active = ?", link.ID, true).Find(&targets)
		
		if len(targets) == 0 {
			if link.BackupURL != "" {
				c.Header("X-GeoIP-Country", location.CountryCode)
				c.Redirect(http.StatusFound, link.BackupURL)
				return
			}
			c.JSON(404, gin.H{"error": "No active targets"})
			return
		}
		
		// Filter targets by country
		var eligibleTargets []Target
		for _, target := range targets {
			if target.Countries == "" || target.Countries == "[]" {
				eligibleTargets = append(eligibleTargets, target)
				continue
			}
			
			var countries []string
			if err := json.Unmarshal([]byte(target.Countries), &countries); err == nil {
				if geoip.IsCountryAllowed(location.CountryCode, countries) {
					eligibleTargets = append(eligibleTargets, target)
				}
			}
		}
		
		if len(eligibleTargets) == 0 {
			c.Header("X-GeoIP-Blocked", "true")
			c.Header("X-GeoIP-Country", location.CountryCode)
			
			if link.BackupURL != "" {
				c.Redirect(http.StatusFound, link.BackupURL)
				return
			}
			c.JSON(403, gin.H{
				"error": "Access denied from your location",
				"country": location.CountryCode,
			})
			return
		}
		
		// Select target (simple: first eligible)
		target := eligibleTargets[0]
		
		// Parse and apply parameter mapping
		var paramMap map[string]string
		if target.ParamMapping != "" {
			json.Unmarshal([]byte(target.ParamMapping), &paramMap)
		}
		
		var staticParams map[string]string
		if target.StaticParams != "" {
			json.Unmarshal([]byte(target.StaticParams), &staticParams)
		}
		
		// Build redirect URL
		redirectURL := target.URL
		queryParams := "?"
		
		// Add query parameters
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
		
		// Add geo info
		queryParams += fmt.Sprintf("geo_country=%s&geo_city=%s&", location.CountryCode, location.City)
		
		if len(queryParams) > 1 {
			redirectURL += queryParams[:len(queryParams)-1]
		}
		
		// Update hit count
		db.Model(&link).Update("current_hits", link.CurrentHits+1)
		db.Model(&target).Update("current_hits", target.CurrentHits+1)
		
		// Set headers
		c.Header("X-Redirect-Info", fmt.Sprintf("Link: %s, Target: %d", linkID, target.ID))
		c.Header("X-GeoIP-Country", location.CountryCode)
		c.Header("X-GeoIP-City", location.City)
		c.Header("X-Client-IP", clientIP)
		
		c.Redirect(http.StatusFound, redirectURL)
	})

	// GeoIP stats endpoint
	v1.GET("/stats/geoip", func(c *gin.Context) {
		var stats []struct {
			CountryCode string `json:"country_code"`
			Count       int    `json:"count"`
		}
		
		// This would normally query from access_logs table
		// For demo, return mock data
		stats = []struct {
			CountryCode string `json:"country_code"`
			Count       int    `json:"count"`
		}{
			{CountryCode: "US", Count: 150},
			{CountryCode: "CN", Count: 80},
			{CountryCode: "UK", Count: 45},
			{CountryCode: "CA", Count: 30},
			{CountryCode: "AU", Count: 25},
		}
		
		c.JSON(200, gin.H{
			"geo_stats": stats,
			"timestamp": time.Now(),
		})
	})

	log.Println("GeoIP-enabled test server starting on :8082")
	log.Println("Available endpoints:")
	log.Println("  GET  /geoip/:ip - Test GeoIP lookup")
	log.Println("  GET  /api/v1/redirect/:bu/:link_id - Redirect with geo filtering")
	log.Println("  GET  /api/v1/stats/geoip - GeoIP statistics")
	
	if err := router.Run(":8082"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}