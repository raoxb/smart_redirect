package geoip

import (
	"net"
	"strings"
)

// IsCountryAllowed checks if the country is in the allowed list
func IsCountryAllowed(countryCode string, allowedCountries []string) bool {
	if len(allowedCountries) == 0 {
		return true
	}

	// Check for "ALL" keyword
	for _, country := range allowedCountries {
		if strings.EqualFold(country, "ALL") {
			return true
		}
	}

	// Check for specific country code
	for _, country := range allowedCountries {
		if strings.EqualFold(countryCode, country) {
			return true
		}
	}

	return false
}

// IsPrivateIP checks if an IP address is private
func IsPrivateIP(ip string) bool {
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