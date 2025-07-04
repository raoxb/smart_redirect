package services

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raoxb/smart_redirect/internal/models"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, func() {
		client.Close()
		mr.Close()
	}
}

func TestIPMemoryService_GetUnusedTarget(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	service := NewIPMemoryService(client)
	ctx := context.Background()

	// Create test targets
	targets := []*models.Target{
		{ID: 1, Weight: 30},
		{ID: 2, Weight: 30},
		{ID: 3, Weight: 40},
	}

	// Test 1: First visit should return first target (no history)
	selected, err := service.GetUnusedTarget(ctx, "192.168.1.1", "test123", targets)
	assert.NoError(t, err)
	assert.NotNil(t, selected)
	assert.Equal(t, uint(1), selected.ID)

	// Test 2: Second visit should return second target (first is marked as visited)
	selected, err = service.GetUnusedTarget(ctx, "192.168.1.1", "test123", targets)
	assert.NoError(t, err)
	assert.NotNil(t, selected)
	assert.Equal(t, uint(2), selected.ID)

	// Test 3: Third visit should return third target
	selected, err = service.GetUnusedTarget(ctx, "192.168.1.1", "test123", targets)
	assert.NoError(t, err)
	assert.NotNil(t, selected)
	assert.Equal(t, uint(3), selected.ID)

	// Test 4: Fourth visit (all visited) should return the one with least visits
	selected, err = service.GetUnusedTarget(ctx, "192.168.1.1", "test123", targets)
	assert.NoError(t, err)
	assert.NotNil(t, selected)
	// Should be any of the targets as they all have 1 visit
}

func TestIPMemoryService_DifferentIPs(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	service := NewIPMemoryService(client)
	ctx := context.Background()

	targets := []*models.Target{
		{ID: 1, Weight: 50},
		{ID: 2, Weight: 50},
	}

	// IP1 visits target 1
	selected1, err := service.GetUnusedTarget(ctx, "192.168.1.1", "link123", targets)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), selected1.ID)

	// IP2 should also get target 1 (different IP, no history)
	selected2, err := service.GetUnusedTarget(ctx, "192.168.1.2", "link123", targets)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), selected2.ID)

	// IP1's second visit should get target 2
	selected3, err := service.GetUnusedTarget(ctx, "192.168.1.1", "link123", targets)
	assert.NoError(t, err)
	assert.Equal(t, uint(2), selected3.ID)
}

func TestIPMemoryService_ClearMemory(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	service := NewIPMemoryService(client)
	ctx := context.Background()

	targets := []*models.Target{
		{ID: 1, Weight: 100},
	}

	// Visit target
	_, err := service.GetUnusedTarget(ctx, "192.168.1.1", "link123", targets)
	assert.NoError(t, err)

	// Clear memory
	err = service.ClearIPMemory(ctx, "192.168.1.1", "link123")
	assert.NoError(t, err)

	// Next visit should get target 1 again (history cleared)
	selected, err := service.GetUnusedTarget(ctx, "192.168.1.1", "link123", targets)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), selected.ID)
}

func TestIPMemoryService_GetIPStats(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	service := NewIPMemoryService(client)
	ctx := context.Background()

	// Create visits for different links
	targets1 := []*models.Target{{ID: 1}, {ID: 2}}
	targets2 := []*models.Target{{ID: 3}, {ID: 4}}

	// Visit link1 targets
	service.GetUnusedTarget(ctx, "192.168.1.1", "link1", targets1)
	service.GetUnusedTarget(ctx, "192.168.1.1", "link1", targets1)

	// Visit link2 targets
	service.GetUnusedTarget(ctx, "192.168.1.1", "link2", targets2)

	// Get stats
	stats, err := service.GetIPStats(ctx, "192.168.1.1")
	assert.NoError(t, err)
	assert.Equal(t, 2, stats["total_links_visited"])

	links := stats["links"].([]map[string]interface{})
	assert.Len(t, links, 2)
}