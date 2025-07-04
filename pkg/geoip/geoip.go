package geoip

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type GeoIP struct {
	client *http.Client
}

type LocationInfo struct {
	IP          string `json:"ip"`
	CountryCode string `json:"country_code"`
	Country     string `json:"country"`
	City        string `json:"city"`
	Region      string `json:"region"`
}

func NewGeoIP() *GeoIP {
	return &GeoIP{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (g *GeoIP) GetLocation(ip string) (*LocationInfo, error) {
	if isPrivateIP(ip) {
		return &LocationInfo{
			IP:          ip,
			CountryCode: "LOCAL",
			Country:     "Local Network",
		}, nil
	}
	
	resp, err := g.client.Get(fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,countryCode,city,region,query", ip))
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
		Status      string `json:"status"`
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
		City        string `json:"city"`
		Region      string `json:"region"`
		Query       string `json:"query"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	
	if result.Status != "success" {
		return nil, fmt.Errorf("api returned error status")
	}
	
	return &LocationInfo{
		IP:          result.Query,
		CountryCode: result.CountryCode,
		Country:     result.Country,
		City:        result.City,
		Region:      result.Region,
	}, nil
}

func (g *GeoIP) IsCountryAllowed(ip string, allowedCountries []string) (bool, error) {
	if len(allowedCountries) == 0 {
		return true, nil
	}
	
	location, err := g.GetLocation(ip)
	if err != nil {
		return false, err
	}
	
	for _, country := range allowedCountries {
		if strings.EqualFold(location.CountryCode, country) {
			return true, nil
		}
	}
	
	return false, nil
}

func isPrivateIP(ip string) bool {
	privateIPBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}
	
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false
	}
	
	for _, block := range privateIPBlocks {
		_, cidr, err := net.ParseCIDR(block)
		if err != nil {
			continue
		}
		if cidr.Contains(ipAddr) {
			return true
		}
	}
	
	return false
}