package main

import (
	"fmt"
	"log"
	"os"

	"github.com/raoxb/smart_redirect/internal/config"
	"github.com/raoxb/smart_redirect/pkg/geoip"
)

func main() {
	// Load config
	cfg, err := config.Load("config/test.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Create GeoIP provider
	provider, err := geoip.NewProvider(&cfg.GeoIP)
	if err != nil {
		log.Fatal("Failed to create GeoIP provider:", err)
	}
	defer provider.Close()

	// Test IPs
	testIPs := []string{
		"8.8.8.8",        // Google DNS (US)
		"1.1.1.1",        // Cloudflare DNS (US/AU)
		"223.5.5.5",      // Alibaba DNS (CN)
		"208.67.222.222", // OpenDNS (US)
		"192.168.1.1",    // Private IP
		"127.0.0.1",      // Localhost
	}

	fmt.Println("=== GeoIP Test Results ===")
	fmt.Println("Provider:", cfg.GeoIP.Provider)
	fmt.Println()

	for _, ip := range testIPs {
		location, err := provider.GetLocation(ip)
		if err != nil {
			fmt.Printf("❌ %s: Error - %v\n", ip, err)
			continue
		}

		fmt.Printf("✅ %s:\n", ip)
		fmt.Printf("   Country: %s (%s)\n", location.CountryName, location.CountryCode)
		fmt.Printf("   Region: %s (%s)\n", location.RegionName, location.RegionCode)
		fmt.Printf("   City: %s\n", location.City)
		if location.Latitude != 0 || location.Longitude != 0 {
			fmt.Printf("   Location: %.4f, %.4f\n", location.Latitude, location.Longitude)
		}
		if location.ISP != "" {
			fmt.Printf("   ISP: %s\n", location.ISP)
		}
		fmt.Println()
	}

	// Test country filtering
	fmt.Println("=== Country Filter Test ===")
	testCountries := []struct {
		ip       string
		allowed  []string
		expected bool
	}{
		{"8.8.8.8", []string{"US", "CA"}, true},        // US IP, allowed
		{"8.8.8.8", []string{"CN", "JP"}, false},       // US IP, not allowed
		{"223.5.5.5", []string{"CN"}, true},            // CN IP, allowed
		{"1.1.1.1", []string{"ALL"}, true},             // Any IP, ALL allowed
		{"192.168.1.1", []string{"US"}, true},          // Private IP, always allowed
		{"8.8.8.8", []string{}, true},                  // Empty list, all allowed
	}

	for _, test := range testCountries {
		location, err := provider.GetLocation(test.ip)
		if err != nil {
			fmt.Printf("❌ Error getting location for %s: %v\n", test.ip, err)
			continue
		}

		allowed := geoip.IsCountryAllowed(location.CountryCode, test.allowed)
		status := "❌"
		if allowed == test.expected {
			status = "✅"
		}

		fmt.Printf("%s IP: %s, Country: %s, Allowed: %v, Expected: %v\n",
			status, test.ip, location.CountryCode, test.allowed, test.expected)
	}

	// Download database if using MaxMind
	if cfg.GeoIP.Provider == "maxmind" {
		fmt.Println("\n=== MaxMind Database Info ===")
		if _, err := os.Stat(cfg.GeoIP.DatabasePath); os.IsNotExist(err) {
			fmt.Println("Database file does not exist. It will be downloaded on first use.")
		} else {
			info, _ := os.Stat(cfg.GeoIP.DatabasePath)
			fmt.Printf("Database file: %s\n", cfg.GeoIP.DatabasePath)
			fmt.Printf("Size: %.2f MB\n", float64(info.Size())/1024/1024)
			fmt.Printf("Modified: %s\n", info.ModTime())
		}
	}
}