package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/raoxb/smart_redirect/internal/models"
)

type IPMemoryService struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewIPMemoryService(redisClient *redis.Client) *IPMemoryService {
	return &IPMemoryService{
		redisClient: redisClient,
		ttl:         12 * time.Hour, // 12 hours memory
	}
}

// GetUnusedTarget returns the first target that the IP hasn't visited
// If all targets have been visited, returns the target with least visits
func (s *IPMemoryService) GetUnusedTarget(ctx context.Context, clientIP string, linkID string, eligibleTargets []*models.Target) (*models.Target, error) {
	if len(eligibleTargets) == 0 {
		return nil, fmt.Errorf("no eligible targets")
	}

	// Key for storing IP memory
	memoryKey := fmt.Sprintf("ip_memory:%s:%s", clientIP, linkID)

	// Get IP's visit history
	visitedTargets, err := s.getVisitHistory(ctx, memoryKey)
	if err != nil {
		// On error, fall back to first target
		return eligibleTargets[0], nil
	}

	// Find first unvisited target
	for _, target := range eligibleTargets {
		targetIDStr := fmt.Sprintf("%d", target.ID)
		if _, visited := visitedTargets[targetIDStr]; !visited {
			// Mark this target as visited
			s.markTargetVisited(ctx, memoryKey, target.ID)
			return target, nil
		}
	}

	// All targets have been visited, find the one with least visits
	var selectedTarget *models.Target
	minVisits := int(^uint(0) >> 1) // Max int

	for _, target := range eligibleTargets {
		targetIDStr := fmt.Sprintf("%d", target.ID)
		visits := visitedTargets[targetIDStr]
		if visits < minVisits {
			minVisits = visits
			selectedTarget = target
		}
	}

	// Mark selected target as visited again
	if selectedTarget != nil {
		s.markTargetVisited(ctx, memoryKey, selectedTarget.ID)
	}

	return selectedTarget, nil
}

// getVisitHistory retrieves the visit count for each target
func (s *IPMemoryService) getVisitHistory(ctx context.Context, key string) (map[string]int, error) {
	data, err := s.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// No history yet
		return make(map[string]int), nil
	}
	if err != nil {
		return nil, err
	}

	var history map[string]int
	if err := json.Unmarshal([]byte(data), &history); err != nil {
		return nil, err
	}

	return history, nil
}

// markTargetVisited increments the visit count for a target
func (s *IPMemoryService) markTargetVisited(ctx context.Context, key string, targetID uint) error {
	// Get current history
	history, err := s.getVisitHistory(ctx, key)
	if err != nil {
		history = make(map[string]int)
	}

	// Increment visit count
	targetIDStr := fmt.Sprintf("%d", targetID)
	history[targetIDStr]++

	// Save back to Redis
	data, err := json.Marshal(history)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, data, s.ttl).Err()
}

// ClearIPMemory clears the visit history for a specific IP and link
func (s *IPMemoryService) ClearIPMemory(ctx context.Context, clientIP string, linkID string) error {
	key := fmt.Sprintf("ip_memory:%s:%s", clientIP, linkID)
	return s.redisClient.Del(ctx, key).Err()
}

// GetIPStats returns statistics about IP visit patterns
func (s *IPMemoryService) GetIPStats(ctx context.Context, clientIP string) (map[string]interface{}, error) {
	pattern := fmt.Sprintf("ip_memory:%s:*", clientIP)
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_links_visited": len(keys),
		"links":               []map[string]interface{}{},
	}

	for _, key := range keys {
		history, err := s.getVisitHistory(ctx, key)
		if err != nil {
			continue
		}

		totalVisits := 0
		for _, visits := range history {
			totalVisits += visits
		}

		linkStats := map[string]interface{}{
			"link_key":      key,
			"targets_count": len(history),
			"total_visits":  totalVisits,
		}
		stats["links"] = append(stats["links"].([]map[string]interface{}), linkStats)
	}

	return stats, nil
}