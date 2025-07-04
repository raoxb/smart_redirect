package middleware

import (
	"net/http"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	
	"github.com/raoxb/smart_redirect/internal/services"
)

func RateLimitMiddleware(redis *redis.Client, limit int, duration time.Duration) gin.HandlerFunc {
	rateLimiter := services.NewRateLimiter(redis)
	
	return func(c *gin.Context) {
		ip := getClientIP(c)
		
		blocked, reason := rateLimiter.IsIPBlocked(ip)
		if blocked {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  "IP blocked",
				"reason": reason,
			})
			c.Abort()
			return
		}
		
		allowed, err := rateLimiter.CheckIPLimit(ip, limit, duration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "rate limit check failed"})
			c.Abort()
			return
		}
		
		if !allowed {
			go rateLimiter.BlockIP(ip, "rate limit exceeded", 24*time.Hour)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

func getClientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		return ip
	}
	
	return c.ClientIP()
}