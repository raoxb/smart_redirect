package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/go-redis/redis/v8"
)

type RateLimiter struct {
	redis *redis.Client
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis: redis,
	}
}

func (r *RateLimiter) CheckIPLimit(ip string, limit int, duration time.Duration) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:ip:%s", ip)
	
	pipe := r.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, duration)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}
	
	count := incr.Val()
	return count <= int64(limit), nil
}

func (r *RateLimiter) CheckIPLinkLimit(ip string, linkID uint, limit int, duration time.Duration) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("rate_limit:ip:%s:link:%d", ip, linkID)
	
	pipe := r.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, duration)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check IP link limit: %w", err)
	}
	
	count := incr.Val()
	return count <= int64(limit), nil
}

func (r *RateLimiter) CheckGlobalCap(key string, cap int) (bool, error) {
	if cap <= 0 {
		return true, nil
	}
	
	ctx := context.Background()
	
	count, err := r.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		count = 0
	} else if err != nil {
		return false, fmt.Errorf("failed to get cap count: %w", err)
	}
	
	return count < int64(cap), nil
}

func (r *RateLimiter) IncrementCap(key string) error {
	ctx := context.Background()
	
	return r.redis.Incr(ctx, key).Err()
}

func (r *RateLimiter) GetCount(key string) (int64, error) {
	ctx := context.Background()
	
	count, err := r.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}
	
	return count, nil
}

func (r *RateLimiter) ResetDailyCounter(key string) error {
	ctx := context.Background()
	
	pipe := r.redis.Pipeline()
	pipe.Set(ctx, key, 0, 24*time.Hour)
	
	_, err := pipe.Exec(ctx)
	return err
}

type IPAccessInfo struct {
	Count       int       `json:"count"`
	LastAccess  time.Time `json:"last_access"`
	Country     string    `json:"country"`
	BlockReason string    `json:"block_reason,omitempty"`
}

func (r *RateLimiter) RecordIPAccess(ip string, country string) error {
	ctx := context.Background()
	key := fmt.Sprintf("ip_access:%s", ip)
	
	info := IPAccessInfo{
		Count:      1,
		LastAccess: time.Now(),
		Country:    country,
	}
	
	existing, err := r.redis.Get(ctx, key).Result()
	if err == nil {
		var existingInfo IPAccessInfo
		if err := json.Unmarshal([]byte(existing), &existingInfo); err == nil {
			info.Count = existingInfo.Count + 1
		}
	}
	
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	
	return r.redis.Set(ctx, key, data, 24*time.Hour).Err()
}

func (r *RateLimiter) GetIPAccessInfo(ip string) (*IPAccessInfo, error) {
	ctx := context.Background()
	key := fmt.Sprintf("ip_access:%s", ip)
	
	data, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	
	var info IPAccessInfo
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		return nil, err
	}
	
	return &info, nil
}

func (r *RateLimiter) BlockIP(ip string, reason string, duration time.Duration) error {
	ctx := context.Background()
	key := fmt.Sprintf("blocked_ip:%s", ip)
	
	return r.redis.Set(ctx, key, reason, duration).Err()
}

func (r *RateLimiter) IsIPBlocked(ip string) (bool, string) {
	ctx := context.Background()
	key := fmt.Sprintf("blocked_ip:%s", ip)
	
	reason, err := r.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, ""
	}
	
	return true, reason
}