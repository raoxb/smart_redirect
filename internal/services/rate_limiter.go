package services

import (
	"context"
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

func (r *RateLimiter) CheckGlobalCap(key string, cap int) (bool, error) {
	ctx := context.Background()
	
	count, err := r.redis.Incr(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check global cap: %w", err)
	}
	
	return count <= int64(cap), nil
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