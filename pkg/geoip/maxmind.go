package geoip

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type MaxMindProvider struct {
	licenseKey string
	dbPath     string
	db         *geoip2.Reader
	cache      map[string]*Location
	cacheMu    sync.RWMutex
	cacheSize  int
}

func NewMaxMindProvider(licenseKey, dbPath string, cacheSize int) (*MaxMindProvider, error) {
	provider := &MaxMindProvider{
		licenseKey: licenseKey,
		dbPath:     dbPath,
		cache:      make(map[string]*Location),
		cacheSize:  cacheSize,
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Download database if it doesn't exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if err := provider.UpdateDatabase(); err != nil {
			return nil, fmt.Errorf("failed to download database: %w", err)
		}
	}

	// Open database
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	provider.db = db

	return provider, nil
}

func (m *MaxMindProvider) GetLocation(ip string) (*Location, error) {
	// Check cache
	m.cacheMu.RLock()
	if loc, exists := m.cache[ip]; exists {
		m.cacheMu.RUnlock()
		return loc, nil
	}
	m.cacheMu.RUnlock()

	// Parse IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Query database
	record, err := m.db.City(parsedIP)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	// Create location
	location := &Location{
		IP:          ip,
		CountryCode: record.Country.IsoCode,
		CountryName: record.Country.Names["en"],
		RegionCode:  "",
		RegionName:  "",
		City:        record.City.Names["en"],
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		TimeZone:    record.Location.TimeZone,
		ISP:         "", // Not available in GeoLite2
		Timestamp:   time.Now(),
	}
	
	// Add region info if available
	if len(record.Subdivisions) > 0 {
		location.RegionCode = record.Subdivisions[0].IsoCode
		location.RegionName = record.Subdivisions[0].Names["en"]
	}

	// Update cache
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()
	
	// Simple cache eviction
	if len(m.cache) >= m.cacheSize {
		// Remove a random entry
		for k := range m.cache {
			delete(m.cache, k)
			break
		}
	}
	m.cache[ip] = location

	return location, nil
}

func (m *MaxMindProvider) UpdateDatabase() error {
	// MaxMind download URL for GeoLite2-City
	url := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=%s&suffix=tar.gz", m.licenseKey)

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download database: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download database: status %d", resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp(filepath.Dir(m.dbPath), "geolite2-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Save to temporary file
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save database: %w", err)
	}
	tmpFile.Close()

	// Extract tar.gz
	if err := m.extractDatabase(tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to extract database: %w", err)
	}

	return nil
}

func (m *MaxMindProvider) extractDatabase(tarPath string) error {
	// Open tar.gz file
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Find and extract .mmdb file
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Look for .mmdb file
		if filepath.Ext(header.Name) == ".mmdb" && filepath.Base(header.Name) == "GeoLite2-City.mmdb" {
			// Create output file
			outFile, err := os.Create(m.dbPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			// Copy file content
			_, err = io.Copy(outFile, tr)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return fmt.Errorf("GeoLite2-City.mmdb not found in archive")
}

func (m *MaxMindProvider) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}