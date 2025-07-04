package unit

import (
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/raoxb/smart_redirect/internal/services"
	"github.com/raoxb/smart_redirect/test/testutil"
)

func TestRateLimiter_CheckIPLimit(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	rateLimiter := services.NewRateLimiter(ts.Redis)
	
	t.Run("IP limit not exceeded", func(t *testing.T) {
		allowed, err := rateLimiter.CheckIPLimit("192.168.1.1", 5, time.Hour)
		require.NoError(t, err)
		assert.True(t, allowed)
		
		// Second request should also be allowed
		allowed, err = rateLimiter.CheckIPLimit("192.168.1.1", 5, time.Hour)
		require.NoError(t, err)
		assert.True(t, allowed)
	})
	
	t.Run("IP limit exceeded", func(t *testing.T) {
		ip := "192.168.1.2"
		
		// Make 5 requests (should all be allowed)
		for i := 0; i < 5; i++ {
			allowed, err := rateLimiter.CheckIPLimit(ip, 5, time.Hour)
			require.NoError(t, err)
			assert.True(t, allowed)
		}
		
		// 6th request should be denied
		allowed, err := rateLimiter.CheckIPLimit(ip, 5, time.Hour)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
}

func TestRateLimiter_CheckIPLinkLimit(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	rateLimiter := services.NewRateLimiter(ts.Redis)
	
	t.Run("IP link limit not exceeded", func(t *testing.T) {
		allowed, err := rateLimiter.CheckIPLinkLimit("192.168.1.1", 1, 3, 12*time.Hour)
		require.NoError(t, err)
		assert.True(t, allowed)
	})
	
	t.Run("IP link limit exceeded", func(t *testing.T) {
		ip := "192.168.1.3"
		linkID := uint(1)
		
		// Make 3 requests (should all be allowed)
		for i := 0; i < 3; i++ {
			allowed, err := rateLimiter.CheckIPLinkLimit(ip, linkID, 3, 12*time.Hour)
			require.NoError(t, err)
			assert.True(t, allowed)
		}
		
		// 4th request should be denied
		allowed, err := rateLimiter.CheckIPLinkLimit(ip, linkID, 3, 12*time.Hour)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
	
	t.Run("Different links have separate limits", func(t *testing.T) {
		ip := "192.168.1.4"
		
		// Exhaust limit for link 1
		for i := 0; i < 3; i++ {
			allowed, err := rateLimiter.CheckIPLinkLimit(ip, 1, 3, 12*time.Hour)
			require.NoError(t, err)
			assert.True(t, allowed)
		}
		
		// Link 1 should be blocked
		allowed, err := rateLimiter.CheckIPLinkLimit(ip, 1, 3, 12*time.Hour)
		require.NoError(t, err)
		assert.False(t, allowed)
		
		// Link 2 should still be allowed
		allowed, err = rateLimiter.CheckIPLinkLimit(ip, 2, 3, 12*time.Hour)
		require.NoError(t, err)
		assert.True(t, allowed)
	})
}

func TestRateLimiter_CheckGlobalCap(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	rateLimiter := services.NewRateLimiter(ts.Redis)
	
	t.Run("Global cap not exceeded", func(t *testing.T) {
		key := "test:cap:1"
		
		allowed, err := rateLimiter.CheckGlobalCap(key, 5)
		require.NoError(t, err)
		assert.True(t, allowed)
	})
	
	t.Run("Global cap exceeded", func(t *testing.T) {
		key := "test:cap:2"
		
		// Increment to reach cap
		for i := 0; i < 5; i++ {
			err := rateLimiter.IncrementCap(key)
			require.NoError(t, err)
		}
		
		// Should be blocked now
		allowed, err := rateLimiter.CheckGlobalCap(key, 5)
		require.NoError(t, err)
		assert.False(t, allowed)
	})
	
	t.Run("Zero cap means unlimited", func(t *testing.T) {
		key := "test:cap:3"
		
		// Even after incrementing, should still be allowed with cap=0
		for i := 0; i < 100; i++ {
			err := rateLimiter.IncrementCap(key)
			require.NoError(t, err)
		}
		
		allowed, err := rateLimiter.CheckGlobalCap(key, 0)
		require.NoError(t, err)
		assert.True(t, allowed)
	})
}

func TestRateLimiter_BlockIP(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	rateLimiter := services.NewRateLimiter(ts.Redis)
	
	t.Run("Block and check IP", func(t *testing.T) {
		ip := "192.168.1.100"
		reason := "rate limit exceeded"
		
		// IP should not be blocked initially
		blocked, _ := rateLimiter.IsIPBlocked(ip)
		assert.False(t, blocked)
		
		// Block the IP
		err := rateLimiter.BlockIP(ip, reason, time.Hour)
		require.NoError(t, err)
		
		// IP should now be blocked
		blocked, returnedReason := rateLimiter.IsIPBlocked(ip)
		assert.True(t, blocked)
		assert.Equal(t, reason, returnedReason)
	})
	
	t.Run("Unblock IP", func(t *testing.T) {
		ip := "192.168.1.101"
		
		// Block the IP first
		err := rateLimiter.BlockIP(ip, "test", time.Hour)
		require.NoError(t, err)
		
		blocked, _ := rateLimiter.IsIPBlocked(ip)
		assert.True(t, blocked)
		
		// Unblock by setting duration to 0
		err = rateLimiter.BlockIP(ip, "", 0)
		require.NoError(t, err)
		
		blocked, _ = rateLimiter.IsIPBlocked(ip)
		assert.False(t, blocked)
	})
}

func TestRateLimiter_RecordIPAccess(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	rateLimiter := services.NewRateLimiter(ts.Redis)
	
	t.Run("Record and retrieve IP access", func(t *testing.T) {
		ip := "192.168.1.200"
		country := "US"
		
		// Record access
		err := rateLimiter.RecordIPAccess(ip, country)
		require.NoError(t, err)
		
		// Retrieve info
		info, err := rateLimiter.GetIPAccessInfo(ip)
		require.NoError(t, err)
		require.NotNil(t, info)
		
		assert.Equal(t, 1, info.Count)
		assert.Equal(t, country, info.Country)
		assert.True(t, time.Since(info.LastAccess) < time.Minute)
	})
	
	t.Run("Multiple accesses increment count", func(t *testing.T) {
		ip := "192.168.1.201"
		country := "CA"
		
		// Record multiple accesses
		for i := 0; i < 3; i++ {
			err := rateLimiter.RecordIPAccess(ip, country)
			require.NoError(t, err)
		}
		
		// Check count
		info, err := rateLimiter.GetIPAccessInfo(ip)
		require.NoError(t, err)
		require.NotNil(t, info)
		
		assert.Equal(t, 3, info.Count)
		assert.Equal(t, country, info.Country)
	})
	
	t.Run("Non-existent IP returns nil", func(t *testing.T) {
		info, err := rateLimiter.GetIPAccessInfo("192.168.1.999")
		require.NoError(t, err)
		assert.Nil(t, info)
	})
}