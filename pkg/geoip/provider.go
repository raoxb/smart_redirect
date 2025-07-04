package geoip

import (
	"fmt"
	"time"

	"github.com/raoxb/smart_redirect/internal/config"
)

// Location represents geographic location information
type Location struct {
	IP          string    `json:"ip"`
	CountryCode string    `json:"country_code"`
	CountryName string    `json:"country_name"`
	RegionCode  string    `json:"region_code"`
	RegionName  string    `json:"region_name"`
	City        string    `json:"city"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	TimeZone    string    `json:"timezone"`
	ISP         string    `json:"isp"`
	Timestamp   time.Time `json:"timestamp"`
}

// Provider is the interface for GeoIP providers
type Provider interface {
	GetLocation(ip string) (*Location, error)
	Close() error
}

// NewProvider creates a new GeoIP provider based on configuration
func NewProvider(cfg *config.GeoIPConfig) (Provider, error) {
	if !cfg.Enabled {
		return &DisabledProvider{}, nil
	}

	switch cfg.Provider {
	case "maxmind":
		if cfg.MaxMindLicenseKey == "" {
			return nil, fmt.Errorf("MaxMind license key is required")
		}
		return NewMaxMindProvider(cfg.MaxMindLicenseKey, cfg.DatabasePath, cfg.CacheSize)
	case "ip-api":
		return NewIPAPIProvider(cfg.CacheSize), nil
	default:
		return nil, fmt.Errorf("unknown GeoIP provider: %s", cfg.Provider)
	}
}

// DisabledProvider is used when GeoIP is disabled
type DisabledProvider struct{}

func (d *DisabledProvider) GetLocation(ip string) (*Location, error) {
	return &Location{
		IP:          ip,
		CountryCode: "XX",
		CountryName: "Unknown",
		Timestamp:   time.Now(),
	}, nil
}

func (d *DisabledProvider) Close() error {
	return nil
}