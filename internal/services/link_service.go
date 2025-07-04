package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
	
	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"gorm.io/gorm"
	
	"github.com/raoxb/smart_redirect/internal/models"
)

type LinkService struct {
	db       *gorm.DB
	redis    *redis.Client
	ipMemory *IPMemoryService
}

func NewLinkService(db *gorm.DB, redis *redis.Client) *LinkService {
	return &LinkService{
		db:       db,
		redis:    redis,
		ipMemory: NewIPMemoryService(redis),
	}
}

func (s *LinkService) CreateLink(link *models.Link) error {
	if link.LinkID == "" {
		link.LinkID = generateLinkID()
	}
	
	if err := s.db.Create(link).Error; err != nil {
		return fmt.Errorf("failed to create link: %w", err)
	}
	
	return s.cacheLink(link)
}

func (s *LinkService) GetLinkByID(linkID string) (*models.Link, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("link:%s", linkID)
	
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var link models.Link
		if err := json.Unmarshal([]byte(cached), &link); err == nil {
			return &link, nil
		}
	}
	
	var link models.Link
	err = s.db.Preload("Targets").Where("link_id = ? AND is_active = ?", linkID, true).First(&link).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get link: %w", err)
	}
	
	_ = s.cacheLink(&link)
	
	return &link, nil
}

func (s *LinkService) SelectTarget(link *models.Link, ip string, country string) (*models.Target, error) {
	if len(link.Targets) == 0 {
		return nil, errors.New("no targets available")
	}
	
	eligibleTargets := make([]*models.Target, 0)
	
	for i := range link.Targets {
		target := &link.Targets[i]
		if !target.IsActive {
			continue
		}
		
		if target.Cap > 0 && target.CurrentHits >= target.Cap {
			continue
		}
		
		if target.Countries != "" && target.Countries != "[]" {
			var allowedCountries []string
			if err := json.Unmarshal([]byte(target.Countries), &allowedCountries); err == nil && len(allowedCountries) > 0 {
				allowed := false
				for _, allowedCountry := range allowedCountries {
					if strings.EqualFold(allowedCountry, country) || strings.EqualFold(allowedCountry, "ALL") {
						allowed = true
						break
					}
				}
				if !allowed {
					continue
				}
			}
		}
		
		eligibleTargets = append(eligibleTargets, target)
	}
	
	if len(eligibleTargets) == 0 {
		return nil, errors.New("no targets available for this country")
	}
	
	// Use IP memory service to select target
	ctx := context.Background()
	selected, err := s.ipMemory.GetUnusedTarget(ctx, ip, link.LinkID, eligibleTargets)
	if err != nil {
		// Fallback to weighted random selection
		totalWeight := 0
		for _, t := range eligibleTargets {
			totalWeight += t.Weight
		}
		selected = selectWeightedRandom(eligibleTargets, totalWeight)
	}
	
	return selected, nil
}

func (s *LinkService) IncrementHits(linkID uint, targetID uint) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Link{}).Where("id = ?", linkID).
			UpdateColumn("current_hits", gorm.Expr("current_hits + ?", 1)).Error; err != nil {
			return err
		}
		
		if err := tx.Model(&models.Target{}).Where("id = ?", targetID).
			UpdateColumn("current_hits", gorm.Expr("current_hits + ?", 1)).Error; err != nil {
			return err
		}
		
		return nil
	})
}

func (s *LinkService) ProcessParameters(target *models.Target, originalParams map[string]string) (map[string]string, error) {
	result := make(map[string]string)
	
	for k, v := range originalParams {
		result[k] = v
	}
	
	if target.ParamMapping != "" {
		var mapping map[string]string
		if err := json.Unmarshal([]byte(target.ParamMapping), &mapping); err == nil {
			for oldKey, newKey := range mapping {
				if val, exists := result[oldKey]; exists {
					result[newKey] = val
					if oldKey != newKey {
						delete(result, oldKey)
					}
				}
			}
		}
	}
	
	if target.StaticParams != "" {
		var staticParams map[string]string
		if err := json.Unmarshal([]byte(target.StaticParams), &staticParams); err == nil {
			for k, v := range staticParams {
				result[k] = v
			}
		}
	}
	
	return result, nil
}

func (s *LinkService) cacheLink(link *models.Link) error {
	ctx := context.Background()
	data, err := json.Marshal(link)
	if err != nil {
		return err
	}
	
	return s.redis.Set(ctx, fmt.Sprintf("link:%s", link.LinkID), data, 1*time.Hour).Err()
}

func generateLinkID() string {
	id := uuid.New().String()
	return strings.ReplaceAll(id[:6], "-", "")
}

func selectWeightedRandom(targets []*models.Target, totalWeight int) *models.Target {
	if len(targets) == 1 {
		return targets[0]
	}
	
	r := rand.Intn(totalWeight)
	for _, target := range targets {
		r -= target.Weight
		if r < 0 {
			return target
		}
	}
	
	return targets[len(targets)-1]
}