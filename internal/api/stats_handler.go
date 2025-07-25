package api

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
)

type StatsHandler struct {
	db           *gorm.DB
	rateLimiter  *services.RateLimiter
	statsService *services.StatsService
}

func NewStatsHandler(db *gorm.DB, redis *redis.Client) *StatsHandler {
	return &StatsHandler{
		db:           db,
		rateLimiter:  services.NewRateLimiter(redis),
		statsService: services.NewStatsService(db, redis),
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
		Where("access_logs.link_id = ?", link.ID).
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

// GetRealtimeStats returns real-time statistics for the dashboard
func (h *StatsHandler) GetRealtimeStats(c *gin.Context) {
	// Get hours parameter (default 24)
	hours := 24
	if hoursStr := c.Query("hours"); hoursStr != "" {
		if h, err := strconv.Atoi(hoursStr); err == nil && h > 0 && h <= 168 {
			hours = h
		}
	}

	stats, err := h.statsService.GetRealtimeStats(hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAccessLogs returns paginated access logs with filtering options
func (h *StatsHandler) GetAccessLogs(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	linkID := c.Query("link_id")
	ip := c.Query("ip")
	country := c.Query("country")
	
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize
	
	// Build query
	query := h.db.Model(&models.AccessLog{})
	
	// Apply filters
	if linkID != "" {
		var link models.Link
		if err := h.db.Where("link_id = ?", linkID).First(&link).Error; err == nil {
			query = query.Where("access_logs.link_id = ?", link.ID)
		}
	}
	if ip != "" {
		query = query.Where("ip ILIKE ?", "%"+ip+"%")
	}
	if country != "" {
		query = query.Where("country = ?", country)
	}
	
	// Get total count
	var total int64
	query.Count(&total)
	
	// Get paginated results
	var logs []models.AccessLog
	query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&logs)
	
	// Manually load associated data
	for i := range logs {
		var link models.Link
		if err := h.db.Where("id = ?", logs[i].LinkID).First(&link).Error; err == nil {
			logs[i].Link = &link
		}
		
		var target models.Target
		if err := h.db.Where("id = ?", logs[i].TargetID).First(&target).Error; err == nil {
			logs[i].Target = &target
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data":       logs,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}