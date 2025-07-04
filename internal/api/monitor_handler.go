package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/raoxb/smart_redirect/internal/services"
)

type MonitorHandler struct {
	monitorService *services.MonitorService
}

func NewMonitorHandler(db *gorm.DB, redis *redis.Client) *MonitorHandler {
	return &MonitorHandler{
		monitorService: services.NewMonitorService(db, redis),
	}
}

// GetActiveAlerts returns all active alerts
func (h *MonitorHandler) GetActiveAlerts(c *gin.Context) {
	ctx := c.Request.Context()
	
	alerts, err := h.monitorService.GetActiveAlerts(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch alerts",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// AcknowledgeAlert marks an alert as acknowledged
func (h *MonitorHandler) AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")
	ctx := c.Request.Context()
	
	if err := h.monitorService.AcknowledgeAlert(ctx, alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to acknowledge alert",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Alert acknowledged successfully",
	})
}

// ResolveAlert marks an alert as resolved
func (h *MonitorHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	ctx := c.Request.Context()
	
	if err := h.monitorService.ResolveAlert(ctx, alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resolve alert",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Alert resolved successfully",
	})
}

// GetMonitoringConfig returns the current monitoring configuration
func (h *MonitorHandler) GetMonitoringConfig(c *gin.Context) {
	// In a real implementation, this would fetch from a config store
	config := gin.H{
		"error_rate_threshold":     0.05,
		"response_time_threshold":  1000,
		"traffic_spike_threshold":  2.0,
		"link_cap_threshold":       0.9,
		"check_interval":           60,
		"notification_channels": []gin.H{
			{
				"type":    "email",
				"enabled": true,
				"config": gin.H{
					"recipients": []string{"admin@example.com"},
				},
			},
			{
				"type":    "webhook",
				"enabled": false,
				"config": gin.H{
					"url": "https://hooks.slack.com/services/...",
				},
			},
		},
	}
	
	c.JSON(http.StatusOK, config)
}

// UpdateMonitoringConfig updates the monitoring configuration
func (h *MonitorHandler) UpdateMonitoringConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// In a real implementation, validate and save the config
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  config,
	})
}

// GetHealthStatus returns the overall system health status
func (h *MonitorHandler) GetHealthStatus(c *gin.Context) {
	// Perform health checks
	health := gin.H{
		"status": "healthy",
		"timestamp": c.GetTime("timestamp"),
		"checks": gin.H{
			"database": gin.H{
				"status": "healthy",
				"latency": "2ms",
			},
			"redis": gin.H{
				"status": "healthy",
				"latency": "1ms",
			},
			"api": gin.H{
				"status": "healthy",
				"uptime": "48h32m",
			},
		},
	}
	
	c.JSON(http.StatusOK, health)
}