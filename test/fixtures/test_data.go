package fixtures

import (
	"time"
	
	"github.com/raoxb/smart_redirect/internal/models"
)

func CreateTestUser() *models.User {
	return &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		Password: "$2a$10$test.hash", // bcrypt hash for "password"
		Role:     "user",
		IsActive: true,
	}
}

func CreateTestAdmin() *models.User {
	return &models.User{
		ID:       2,
		Username: "admin",
		Email:    "admin@example.com",
		Password: "$2a$10$test.hash", // bcrypt hash for "password"
		Role:     "admin",
		IsActive: true,
	}
}

func CreateTestLink() *models.Link {
	return &models.Link{
		ID:           1,
		LinkID:       "abc123",
		BusinessUnit: "bu01",
		Network:      "mi",
		TotalCap:     1000,
		CurrentHits:  100,
		BackupURL:    "https://backup.example.com",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func CreateTestTargets() []models.Target {
	return []models.Target{
		{
			ID:            1,
			LinkID:        1,
			URL:           "https://target1.example.com",
			Weight:        70,
			Cap:           500,
			CurrentHits:   50,
			Countries:     `["US","CA"]`,
			ParamMapping:  `{"kw":"q"}`,
			StaticParams:  `{"ref":"test"}`,
			IsActive:      true,
		},
		{
			ID:            2,
			LinkID:        1,
			URL:           "https://target2.example.com",
			Weight:        30,
			Cap:           300,
			CurrentHits:   25,
			Countries:     `["UK","DE"]`,
			ParamMapping:  `{}`,
			StaticParams:  `{"campaign":"test2"}`,
			IsActive:      true,
		},
	}
}

func CreateTestAccessLog() *models.AccessLog {
	return &models.AccessLog{
		ID:        1,
		LinkID:    1,
		TargetID:  1,
		IP:        "192.168.1.1",
		UserAgent: "Mozilla/5.0 Test Browser",
		Referer:   "https://google.com",
		Country:   "US",
		CreatedAt: time.Now(),
	}
}