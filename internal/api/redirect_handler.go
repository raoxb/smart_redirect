package api

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
	"github.com/raoxb/smart_redirect/pkg/geoip"
)

type RedirectHandler struct {
	linkService *services.LinkService
	rateLimiter *services.RateLimiter
	geoIP       *geoip.GeoIP
	db          *gorm.DB
}

func NewRedirectHandler(db *gorm.DB, redis *redis.Client) *RedirectHandler {
	return &RedirectHandler{
		linkService: services.NewLinkService(db, redis),
		rateLimiter: services.NewRateLimiter(redis),
		geoIP:       geoip.NewGeoIP(),
		db:          db,
	}
}

func (h *RedirectHandler) HandleRedirect(c *gin.Context) {
	bu := c.Param("bu")
	linkID := c.Param("link_id")
	clientIP := getClientIP(c)
	
	blocked, reason := h.rateLimiter.IsIPBlocked(clientIP)
	if blocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "IP blocked", "reason": reason})
		return
	}
	
	location, err := h.geoIP.GetLocation(clientIP)
	if err != nil {
		location = &geoip.LocationInfo{
			IP:          clientIP,
			CountryCode: "UNKNOWN",
			Country:     "Unknown",
		}
	}
	
	link, err := h.linkService.GetLinkByID(linkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	
	if link == nil || link.BusinessUnit != bu {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	allowed, err := h.rateLimiter.CheckIPLimit(clientIP, 100, time.Hour)
	if err != nil || !allowed {
		go h.rateLimiter.BlockIP(clientIP, "rate limit exceeded", 24*time.Hour)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		return
	}
	
	allowed, err = h.rateLimiter.CheckIPLinkLimit(clientIP, link.ID, 10, 12*time.Hour)
	if err != nil || !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "link access limit exceeded"})
		return
	}
	
	globalCapKey := fmt.Sprintf("global_cap:link:%d", link.ID)
	allowed, err = h.rateLimiter.CheckGlobalCap(globalCapKey, link.TotalCap)
	if err != nil || !allowed {
		if link.BackupURL != "" {
			c.Redirect(http.StatusFound, link.BackupURL)
			return
		}
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "link cap reached"})
		return
	}
	
	target, err := h.linkService.SelectTarget(link, clientIP, location.CountryCode)
	if err != nil {
		if link.BackupURL != "" {
			c.Redirect(http.StatusFound, link.BackupURL)
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	
	originalParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			originalParams[key] = values[0]
		}
	}
	
	processedParams, err := h.linkService.ProcessParameters(target, originalParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process parameters"})
		return
	}
	
	targetURL, err := url.Parse(target.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid target URL"})
		return
	}
	
	query := targetURL.Query()
	for k, v := range processedParams {
		query.Set(k, v)
	}
	targetURL.RawQuery = query.Encode()
	
	go func() {
		_ = h.linkService.IncrementHits(link.ID, target.ID)
		_ = h.rateLimiter.IncrementCap(globalCapKey)
		_ = h.rateLimiter.RecordIPAccess(clientIP, location.CountryCode)
		
		accessLog := &models.AccessLog{
			LinkID:    link.ID,
			TargetID:  target.ID,
			IP:        clientIP,
			UserAgent: c.GetHeader("User-Agent"),
			Referer:   c.GetHeader("Referer"),
			Country:   location.CountryCode,
		}
		_ = h.db.Create(accessLog)
	}()
	
	c.Redirect(http.StatusFound, targetURL.String())
}

func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	
	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return ip
	}
	
	return c.Request.RemoteAddr
}