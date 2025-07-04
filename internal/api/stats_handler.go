package api

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
)

type StatsHandler struct {
	db          *gorm.DB
	rateLimiter *services.RateLimiter
}

func NewStatsHandler(db *gorm.DB, redis *redis.Client) *StatsHandler {
	return &StatsHandler{
		db:          db,
		rateLimiter: services.NewRateLimiter(redis),
	}
}

type LinkStats struct {
	LinkID       string `json:"link_id"`
	BusinessUnit string `json:"business_unit"`
	TotalHits    int64  `json:"total_hits"`
	TodayHits    int64  `json:"today_hits"`
	UniqueIPs    int64  `json:"unique_ips"`
	Countries    []CountryStats `json:"countries"`
	Targets      []TargetStats  `json:"targets"`
}

type CountryStats struct {
	Country string `json:"country"`
	Hits    int64  `json:"hits"`
}

type TargetStats struct {
	TargetID uint   `json:"target_id"`
	URL      string `json:"url"`
	Hits     int64  `json:"hits"`
}

func (h *StatsHandler) GetLinkStats(c *gin.Context) {
	linkID := c.Param("link_id")
	
	var link models.Link
	if err := h.db.Where("link_id = ?", linkID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	var totalHits int64
	h.db.Model(&models.AccessLog{}).Where("link_id = ?", link.ID).Count(&totalHits)
	
	today := time.Now().Truncate(24 * time.Hour)
	var todayHits int64
	h.db.Model(&models.AccessLog{}).Where("link_id = ? AND created_at >= ?", link.ID, today).Count(&todayHits)
	
	var uniqueIPs int64
	h.db.Model(&models.AccessLog{}).Where("link_id = ?", link.ID).Distinct("ip").Count(&uniqueIPs)
	
	var countries []CountryStats
	h.db.Model(&models.AccessLog{}).
		Where("link_id = ?", link.ID).
		Select("country, COUNT(*) as hits").
		Group("country").
		Scan(&countries)
	
	var targets []TargetStats
	h.db.Model(&models.AccessLog{}).
		Joins("JOIN targets ON access_logs.target_id = targets.id").
		Where("link_id = ?", link.ID).
		Select("targets.id as target_id, targets.url, COUNT(*) as hits").
		Group("targets.id, targets.url").
		Scan(&targets)
	
	stats := LinkStats{
		LinkID:       link.LinkID,
		BusinessUnit: link.BusinessUnit,
		TotalHits:    totalHits,
		TodayHits:    todayHits,
		UniqueIPs:    uniqueIPs,
		Countries:    countries,
		Targets:      targets,
	}
	
	c.JSON(http.StatusOK, stats)
}

func (h *StatsHandler) GetSystemStats(c *gin.Context) {
	var totalLinks int64
	h.db.Model(&models.Link{}).Count(&totalLinks)
	
	var totalHits int64
	h.db.Model(&models.AccessLog{}).Count(&totalHits)
	
	today := time.Now().Truncate(24 * time.Hour)
	var todayHits int64
	h.db.Model(&models.AccessLog{}).Where("created_at >= ?", today).Count(&todayHits)
	
	var uniqueIPs int64
	h.db.Model(&models.AccessLog{}).Distinct("ip").Count(&uniqueIPs)
	
	var topCountries []CountryStats
	h.db.Model(&models.AccessLog{}).
		Select("country, COUNT(*) as hits").
		Group("country").
		Order("hits DESC").
		Limit(10).
		Scan(&topCountries)
	
	c.JSON(http.StatusOK, gin.H{
		"total_links":   totalLinks,
		"total_hits":    totalHits,
		"today_hits":    todayHits,
		"unique_ips":    uniqueIPs,
		"top_countries": topCountries,
	})
}

func (h *StatsHandler) GetIPInfo(c *gin.Context) {
	ip := c.Param("ip")
	
	info, err := h.rateLimiter.GetIPAccessInfo(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get IP info"})
		return
	}
	
	if info == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "IP not found"})
		return
	}
	
	blocked, reason := h.rateLimiter.IsIPBlocked(ip)
	
	var accessLogs []models.AccessLog
	h.db.Where("ip = ?", ip).
		Order("created_at DESC").
		Limit(20).
		Preload("Link").
		Preload("Target").
		Find(&accessLogs)
	
	c.JSON(http.StatusOK, gin.H{
		"ip":           ip,
		"access_count": info.Count,
		"last_access":  info.LastAccess,
		"country":      info.Country,
		"is_blocked":   blocked,
		"block_reason": reason,
		"recent_logs":  accessLogs,
	})
}

func (h *StatsHandler) BlockIP(c *gin.Context) {
	ip := c.Param("ip")
	
	type BlockRequest struct {
		Reason   string `json:"reason" binding:"required"`
		Duration int    `json:"duration"` // hours
	}
	
	var req BlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	duration := time.Duration(req.Duration) * time.Hour
	if duration == 0 {
		duration = 24 * time.Hour
	}
	
	if err := h.rateLimiter.BlockIP(ip, req.Reason, duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block IP"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "IP blocked successfully"})
}

func (h *StatsHandler) UnblockIP(c *gin.Context) {
	ip := c.Param("ip")
	
	if err := h.rateLimiter.BlockIP(ip, "", 0); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock IP"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "IP unblocked successfully"})
}

func (h *StatsHandler) GetHourlyStats(c *gin.Context) {
	linkID := c.Param("link_id")
	hoursStr := c.DefaultQuery("hours", "24")
	hours, _ := strconv.Atoi(hoursStr)
	
	if hours > 168 { // max 7 days
		hours = 168
	}
	
	var link models.Link
	if err := h.db.Where("link_id = ?", linkID).First(&link).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	type HourlyStats struct {
		Hour string `json:"hour"`
		Hits int64  `json:"hits"`
	}
	
	var stats []HourlyStats
	h.db.Model(&models.AccessLog{}).
		Where("link_id = ? AND created_at >= ?", link.ID, time.Now().Add(-time.Duration(hours)*time.Hour)).
		Select("DATE_TRUNC('hour', created_at) as hour, COUNT(*) as hits").
		Group("hour").
		Order("hour").
		Scan(&stats)
	
	c.JSON(http.StatusOK, stats)
}