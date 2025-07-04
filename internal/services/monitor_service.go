package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/raoxb/smart_redirect/internal/models"
)

type MonitorService struct {
	db          *gorm.DB
	redis       *redis.Client
	alertConfig AlertConfig
}

type AlertConfig struct {
	ErrorRateThreshold    float64       `json:"error_rate_threshold"`
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"`
	TrafficSpikeThreshold float64       `json:"traffic_spike_threshold"`
	LinkCapThreshold      float64       `json:"link_cap_threshold"`
	CheckInterval         time.Duration `json:"check_interval"`
}

type Alert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Level       string    `json:"level"` // info, warning, critical
	Title       string    `json:"title"`
	Message     string    `json:"message"`
	Details     map[string]interface{} `json:"details"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	Acknowledged bool     `json:"acknowledged"`
}

func NewMonitorService(db *gorm.DB, redis *redis.Client) *MonitorService {
	return &MonitorService{
		db:    db,
		redis: redis,
		alertConfig: AlertConfig{
			ErrorRateThreshold:    0.05,  // 5% error rate
			ResponseTimeThreshold: 1000,  // 1000ms
			TrafficSpikeThreshold: 2.0,   // 2x normal traffic
			LinkCapThreshold:      0.9,   // 90% of cap
			CheckInterval:         60,    // 60 seconds
		},
	}
}

// StartMonitoring starts the background monitoring process
func (s *MonitorService) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(s.alertConfig.CheckInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runChecks(ctx)
		}
	}
}

func (s *MonitorService) runChecks(ctx context.Context) {
	// Check error rates
	s.checkErrorRates(ctx)
	
	// Check response times
	s.checkResponseTimes(ctx)
	
	// Check traffic patterns
	s.checkTrafficPatterns(ctx)
	
	// Check link capacities
	s.checkLinkCapacities(ctx)
	
	// Check system health
	s.checkSystemHealth(ctx)
}

func (s *MonitorService) checkErrorRates(ctx context.Context) {
	// Check 5xx errors in the last 5 minutes
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
	
	var totalRequests int64
	var errorRequests int64
	
	// In a real implementation, you'd track HTTP status codes
	// For now, we'll check access logs
	s.db.Model(&models.AccessLog{}).
		Where("created_at >= ?", fiveMinutesAgo).
		Count(&totalRequests)
	
	if totalRequests > 100 { // Only alert if we have enough data
		errorRate := float64(errorRequests) / float64(totalRequests)
		if errorRate > s.alertConfig.ErrorRateThreshold {
			s.createAlert(ctx, &Alert{
				Type:  "error_rate",
				Level: "critical",
				Title: "High Error Rate Detected",
				Message: fmt.Sprintf("Error rate is %.2f%% (threshold: %.2f%%)", 
					errorRate*100, s.alertConfig.ErrorRateThreshold*100),
				Details: map[string]interface{}{
					"error_count": errorRequests,
					"total_count": totalRequests,
					"error_rate":  errorRate,
				},
			})
		}
	}
}

func (s *MonitorService) checkResponseTimes(ctx context.Context) {
	// Check average response time
	key := fmt.Sprintf("stats:response_time:%s", time.Now().Format("2006-01-02:15:04"))
	avgTimeStr, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var avgTime float64
		if json.Unmarshal([]byte(avgTimeStr), &avgTime) == nil {
			if avgTime > float64(s.alertConfig.ResponseTimeThreshold) {
				s.createAlert(ctx, &Alert{
					Type:  "response_time",
					Level: "warning",
					Title: "High Response Time",
					Message: fmt.Sprintf("Average response time is %.2fms (threshold: %dms)", 
						avgTime, s.alertConfig.ResponseTimeThreshold),
					Details: map[string]interface{}{
						"avg_response_time": avgTime,
						"threshold":         s.alertConfig.ResponseTimeThreshold,
					},
				})
			}
		}
	}
}

func (s *MonitorService) checkTrafficPatterns(ctx context.Context) {
	// Compare current hour traffic with same hour yesterday
	now := time.Now()
	currentHourKey := fmt.Sprintf("stats:counter:%s", now.Format("2006-01-02:15"))
	yesterdayHourKey := fmt.Sprintf("stats:counter:%s", now.AddDate(0, 0, -1).Format("2006-01-02:15"))
	
	currentCount, _ := s.redis.Get(ctx, currentHourKey).Int64()
	yesterdayCount, _ := s.redis.Get(ctx, yesterdayHourKey).Int64()
	
	if yesterdayCount > 0 {
		spike := float64(currentCount) / float64(yesterdayCount)
		if spike > s.alertConfig.TrafficSpikeThreshold {
			s.createAlert(ctx, &Alert{
				Type:  "traffic_spike",
				Level: "warning",
				Title: "Traffic Spike Detected",
				Message: fmt.Sprintf("Traffic is %.1fx higher than usual", spike),
				Details: map[string]interface{}{
					"current_traffic":   currentCount,
					"yesterday_traffic": yesterdayCount,
					"spike_ratio":       spike,
				},
			})
		}
	}
}

func (s *MonitorService) checkLinkCapacities(ctx context.Context) {
	// Check links approaching their cap
	var links []models.Link
	s.db.Where("is_active = ? AND total_cap > 0", true).Find(&links)
	
	for _, link := range links {
		usagePercent := float64(link.CurrentHits) / float64(link.TotalCap)
		if usagePercent > s.alertConfig.LinkCapThreshold {
			s.createAlert(ctx, &Alert{
				Type:  "link_cap",
				Level: "warning",
				Title: fmt.Sprintf("Link %s Approaching Cap", link.LinkID),
				Message: fmt.Sprintf("Link has used %.1f%% of its cap (%d/%d)", 
					usagePercent*100, link.CurrentHits, link.TotalCap),
				Details: map[string]interface{}{
					"link_id":      link.LinkID,
					"current_hits": link.CurrentHits,
					"total_cap":    link.TotalCap,
					"usage_percent": usagePercent * 100,
				},
			})
		}
	}
}

func (s *MonitorService) checkSystemHealth(ctx context.Context) {
	// Check Redis connectivity
	if err := s.redis.Ping(ctx).Err(); err != nil {
		s.createAlert(ctx, &Alert{
			Type:    "system_health",
			Level:   "critical",
			Title:   "Redis Connection Failed",
			Message: "Unable to connect to Redis",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
	}
	
	// Check database connectivity
	sqlDB, err := s.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		s.createAlert(ctx, &Alert{
			Type:    "system_health",
			Level:   "critical",
			Title:   "Database Connection Failed",
			Message: "Unable to connect to database",
			Details: map[string]interface{}{
				"error": err,
			},
		})
	}
}

func (s *MonitorService) createAlert(ctx context.Context, alert *Alert) {
	alert.ID = fmt.Sprintf("%s_%d", alert.Type, time.Now().Unix())
	alert.CreatedAt = time.Now()
	
	// Store alert in Redis
	alertKey := fmt.Sprintf("alerts:%s", alert.ID)
	data, _ := json.Marshal(alert)
	s.redis.Set(ctx, alertKey, data, 24*time.Hour)
	
	// Add to active alerts set
	s.redis.SAdd(ctx, "alerts:active", alert.ID)
	
	// Log the alert
	log.Printf("[%s] %s: %s", alert.Level, alert.Title, alert.Message)
	
	// In a real implementation, you would also:
	// - Send notifications (email, Slack, etc.)
	// - Update monitoring dashboards
	// - Trigger automated responses
}

// GetActiveAlerts returns all active alerts
func (s *MonitorService) GetActiveAlerts(ctx context.Context) ([]*Alert, error) {
	alertIDs, err := s.redis.SMembers(ctx, "alerts:active").Result()
	if err != nil {
		return nil, err
	}
	
	var alerts []*Alert
	for _, id := range alertIDs {
		alertKey := fmt.Sprintf("alerts:%s", id)
		data, err := s.redis.Get(ctx, alertKey).Result()
		if err != nil {
			continue
		}
		
		var alert Alert
		if json.Unmarshal([]byte(data), &alert) == nil {
			alerts = append(alerts, &alert)
		}
	}
	
	return alerts, nil
}

// AcknowledgeAlert marks an alert as acknowledged
func (s *MonitorService) AcknowledgeAlert(ctx context.Context, alertID string) error {
	alertKey := fmt.Sprintf("alerts:%s", alertID)
	data, err := s.redis.Get(ctx, alertKey).Result()
	if err != nil {
		return err
	}
	
	var alert Alert
	if err := json.Unmarshal([]byte(data), &alert); err != nil {
		return err
	}
	
	alert.Acknowledged = true
	updatedData, _ := json.Marshal(alert)
	return s.redis.Set(ctx, alertKey, updatedData, 24*time.Hour).Err()
}

// ResolveAlert marks an alert as resolved
func (s *MonitorService) ResolveAlert(ctx context.Context, alertID string) error {
	alertKey := fmt.Sprintf("alerts:%s", alertID)
	data, err := s.redis.Get(ctx, alertKey).Result()
	if err != nil {
		return err
	}
	
	var alert Alert
	if err := json.Unmarshal([]byte(data), &alert); err != nil {
		return err
	}
	
	now := time.Now()
	alert.ResolvedAt = &now
	updatedData, _ := json.Marshal(alert)
	
	// Update alert
	s.redis.Set(ctx, alertKey, updatedData, 24*time.Hour)
	
	// Remove from active alerts
	s.redis.SRem(ctx, "alerts:active", alertID)
	
	return nil
}