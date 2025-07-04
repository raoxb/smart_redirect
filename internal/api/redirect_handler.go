package api

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
)

type RedirectHandler struct {
	linkService *services.LinkService
	rateLimiter *services.RateLimiter
	db          *gorm.DB
}

func NewRedirectHandler(db *gorm.DB, redis *redis.Client) *RedirectHandler {
	return &RedirectHandler{
		linkService: services.NewLinkService(db, redis),
		rateLimiter: services.NewRateLimiter(redis),
		db:          db,
	}
}

func (h *RedirectHandler) HandleRedirect(c *gin.Context) {
	bu := c.Param("bu")
	linkID := c.Param("link_id")
	
	link, err := h.linkService.GetLinkByID(linkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	
	if link == nil || link.BusinessUnit != bu {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}
	
	if link.TotalCap > 0 && link.CurrentHits >= link.TotalCap {
		if link.BackupURL != "" {
			c.Redirect(http.StatusFound, link.BackupURL)
			return
		}
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "link cap reached"})
		return
	}
	
	clientIP := getClientIP(c)
	
	target, err := h.linkService.SelectTarget(link, clientIP)
	if err != nil {
		if link.BackupURL != "" {
			c.Redirect(http.StatusFound, link.BackupURL)
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no available targets"})
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
		
		accessLog := &models.AccessLog{
			LinkID:    link.ID,
			TargetID:  target.ID,
			IP:        clientIP,
			UserAgent: c.GetHeader("User-Agent"),
			Referer:   c.GetHeader("Referer"),
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