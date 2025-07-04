package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/raoxb/smart_redirect/internal/models"
)

type StatsService struct {
	db    *gorm.DB
	redis *redis.Client
}

type HourlyStats struct {
	Hour       string `json:"hour"`
	Visits     int    `json:"visits"`
	UniqueIPs  int    `json:"unique_ips"`
	Redirects  int    `json:"redirects"`
	Blocked    int    `json:"blocked"`
}

type GeoStats struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Count       int    `json:"count"`
	Percentage  float64 `json:"percentage"`
}

type TargetStats struct {
	TargetID    uint    `json:"target_id"`
	URL         string  `json:"url"`
	Hits        int     `json:"hits"`
	Percentage  float64 `json:"percentage"`
}

type RealtimeStats struct {
	Hourly      []HourlyStats  `json:"hourly"`
	Geographic  []GeoStats     `json:"geographic"`
	TopTargets  []TargetStats  `json:"top_targets"`
	Summary     map[string]interface{} `json:"summary"`
	LastUpdated time.Time      `json:"last_updated"`
}

func NewStatsService(db *gorm.DB, redis *redis.Client) *StatsService {
	return &StatsService{
		db:    db,
		redis: redis,
	}
}

// GetRealtimeStats returns comprehensive statistics for the dashboard
func (s *StatsService) GetRealtimeStats(hours int) (*RealtimeStats, error) {
	ctx := context.Background()
	now := time.Now()
	
	stats := &RealtimeStats{
		LastUpdated: now,
		Summary:     make(map[string]interface{}),
	}

	// Get hourly stats
	hourlyStats, err := s.getHourlyStats(ctx, hours)
	if err != nil {
		return nil, err
	}
	stats.Hourly = hourlyStats

	// Get geographic stats
	geoStats, err := s.getGeographicStats(ctx)
	if err != nil {
		return nil, err
	}
	stats.Geographic = geoStats

	// Get top targets
	targetStats, err := s.getTopTargets(ctx, 10)
	if err != nil {
		return nil, err
	}
	stats.TopTargets = targetStats

	// Calculate summary
	stats.Summary = s.calculateSummary(ctx)

	return stats, nil
}

func (s *StatsService) getHourlyStats(ctx context.Context, hours int) ([]HourlyStats, error) {
	var stats []HourlyStats
	now := time.Now()

	for i := hours - 1; i >= 0; i-- {
		hour := now.Add(-time.Duration(i) * time.Hour)
		hourKey := hour.Format("2006-01-02:15")
		
		// Get stats from Redis cache
		cacheKey := fmt.Sprintf("stats:hourly:%s", hourKey)
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var hourStats HourlyStats
			if json.Unmarshal([]byte(cached), &hourStats) == nil {
				stats = append(stats, hourStats)
				continue
			}
		}

		// Calculate from database
		var count int64
		startTime := hour.Truncate(time.Hour)
		endTime := startTime.Add(time.Hour)

		s.db.Model(&models.AccessLog{}).
			Where("created_at >= ? AND created_at < ?", startTime, endTime).
			Count(&count)

		var uniqueIPs int64
		s.db.Model(&models.AccessLog{}).
			Where("created_at >= ? AND created_at < ?", startTime, endTime).
			Distinct("client_ip").
			Count(&uniqueIPs)

		hourStats := HourlyStats{
			Hour:      hour.Format("15:00"),
			Visits:    int(count),
			UniqueIPs: int(uniqueIPs),
			Redirects: int(count), // Simplified - all visits are redirects
			Blocked:   0,          // Would need to track blocked requests
		}

		// Cache the result
		data, _ := json.Marshal(hourStats)
		s.redis.Set(ctx, cacheKey, data, 2*time.Hour)

		stats = append(stats, hourStats)
	}

	return stats, nil
}

func (s *StatsService) getGeographicStats(ctx context.Context) ([]GeoStats, error) {
	var stats []GeoStats
	
	// Get country distribution from access logs
	type countryCount struct {
		Country string
		Count   int64
	}
	
	var countries []countryCount
	s.db.Model(&models.AccessLog{}).
		Select("country as country, count(*) as count").
		Where("created_at >= ?", time.Now().Add(-24*time.Hour)).
		Group("country").
		Order("count DESC").
		Limit(10).
		Scan(&countries)

	// Calculate total for percentages
	var total int64
	for _, c := range countries {
		total += c.Count
	}

	// Convert to GeoStats
	for _, c := range countries {
		percentage := 0.0
		if total > 0 {
			percentage = float64(c.Count) / float64(total) * 100
		}
		
		stats = append(stats, GeoStats{
			CountryCode: c.Country,
			CountryName: getCountryName(c.Country), // Helper function
			Count:       int(c.Count),
			Percentage:  percentage,
		})
	}

	return stats, nil
}

func (s *StatsService) getTopTargets(ctx context.Context, limit int) ([]TargetStats, error) {
	var stats []TargetStats
	
	// Get top targets by hits
	var targets []models.Target
	s.db.Model(&models.Target{}).
		Where("current_hits > 0").
		Order("current_hits DESC").
		Limit(limit).
		Find(&targets)

	// Calculate total hits
	var totalHits int
	for _, t := range targets {
		totalHits += t.CurrentHits
	}

	// Convert to TargetStats
	for _, t := range targets {
		percentage := 0.0
		if totalHits > 0 {
			percentage = float64(t.CurrentHits) / float64(totalHits) * 100
		}
		
		stats = append(stats, TargetStats{
			TargetID:   t.ID,
			URL:        t.URL,
			Hits:       t.CurrentHits,
			Percentage: percentage,
		})
	}

	return stats, nil
}

func (s *StatsService) calculateSummary(ctx context.Context) map[string]interface{} {
	summary := make(map[string]interface{})

	// Total links
	var totalLinks int64
	s.db.Model(&models.Link{}).Count(&totalLinks)
	summary["total_links"] = totalLinks

	// Active links
	var activeLinks int64
	s.db.Model(&models.Link{}).Where("is_active = ?", true).Count(&activeLinks)
	summary["active_links"] = activeLinks

	// Total visits today
	var todayVisits int64
	s.db.Model(&models.AccessLog{}).
		Where("created_at >= ?", time.Now().Truncate(24*time.Hour)).
		Count(&todayVisits)
	summary["today_visits"] = todayVisits

	// Total visits this week
	var weekVisits int64
	s.db.Model(&models.AccessLog{}).
		Where("created_at >= ?", time.Now().AddDate(0, 0, -7)).
		Count(&weekVisits)
	summary["week_visits"] = weekVisits

	// Average response time (mock for now)
	summary["avg_response_time"] = "45ms"

	// Success rate
	summary["success_rate"] = "99.5%"

	return summary
}

// RecordVisit records a visit in real-time stats
func (s *StatsService) RecordVisit(ctx context.Context, linkID string, targetID uint, clientIP string, country string) error {
	// Update hourly counter
	hourKey := time.Now().Format("2006-01-02:15")
	counterKey := fmt.Sprintf("stats:counter:%s", hourKey)
	s.redis.Incr(ctx, counterKey)
	s.redis.Expire(ctx, counterKey, 2*time.Hour)

	// Update unique IPs set
	ipSetKey := fmt.Sprintf("stats:ips:%s", hourKey)
	s.redis.SAdd(ctx, ipSetKey, clientIP)
	s.redis.Expire(ctx, ipSetKey, 2*time.Hour)

	// Update country counter
	if country != "" {
		countryKey := fmt.Sprintf("stats:country:%s", time.Now().Format("2006-01-02"))
		s.redis.HIncrBy(ctx, countryKey, country, 1)
		s.redis.Expire(ctx, countryKey, 25*time.Hour)
	}

	return nil
}

// Helper function to get country name from code
func getCountryName(code string) string {
	countryNames := map[string]string{
		"US": "United States",
		"CN": "China",
		"UK": "United Kingdom",
		"CA": "Canada",
		"AU": "Australia",
		"DE": "Germany",
		"FR": "France",
		"JP": "Japan",
		"BR": "Brazil",
		"IN": "India",
	}
	
	if name, exists := countryNames[code]; exists {
		return name
	}
	return code
}