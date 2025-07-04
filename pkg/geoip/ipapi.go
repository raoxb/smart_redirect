package geoip

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type IPAPIProvider struct {
	client    *http.Client
	cache     map[string]*Location
	cacheMu   sync.RWMutex
	cacheSize int
}

func NewIPAPIProvider(cacheSize int) *IPAPIProvider {
	return &IPAPIProvider{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cache:     make(map[string]*Location),
		cacheSize: cacheSize,
	}
}

func (i *IPAPIProvider) GetLocation(ip string) (*Location, error) {
	// Check cache
	i.cacheMu.RLock()
	if loc, exists := i.cache[ip]; exists {
		i.cacheMu.RUnlock()
		return loc, nil
	}
	i.cacheMu.RUnlock()

	// Check if private IP
	if IsPrivateIP(ip) {
		return &Location{
			IP:          ip,
			CountryCode: "LOCAL",
			CountryName: "Local Network",
			Timestamp:   time.Now(),
		}, nil
	}

	// Query IP-API
	resp, err := i.client.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,regionName,region,city,lat,lon,timezone,isp,query", ip))
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"region"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		ISP         string  `json:"isp"`
		Query       string  `json:"query"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf("api returned error status")
	}

	location := &Location{
		IP:          result.Query,
		CountryCode: result.CountryCode,
		CountryName: result.Country,
		RegionCode:  result.Region,
		RegionName:  result.RegionName,
		City:        result.City,
		Latitude:    result.Lat,
		Longitude:   result.Lon,
		TimeZone:    result.Timezone,
		ISP:         result.ISP,
		Timestamp:   time.Now(),
	}

	// Update cache
	i.cacheMu.Lock()
	defer i.cacheMu.Unlock()

	// Simple cache eviction
	if len(i.cache) >= i.cacheSize {
		// Remove a random entry
		for k := range i.cache {
			delete(i.cache, k)
			break
		}
	}
	i.cache[ip] = location

	return location, nil
}

func (i *IPAPIProvider) Close() error {
	// Nothing to close for IP-API
	return nil
}