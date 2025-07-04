package unit

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
	"github.com/raoxb/smart_redirect/test/fixtures"
	"github.com/raoxb/smart_redirect/test/testutil"
)

func TestLinkService_CreateLink(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	linkService := services.NewLinkService(ts.DB, ts.Redis)
	
	t.Run("Create link successfully", func(t *testing.T) {
		link := &models.Link{
			BusinessUnit: "bu01",
			Network:      "mi",
			TotalCap:     1000,
			BackupURL:    "https://backup.example.com",
		}
		
		err := linkService.CreateLink(link)
		require.NoError(t, err)
		
		assert.NotEmpty(t, link.LinkID)
		assert.Equal(t, 6, len(link.LinkID))
		assert.NotZero(t, link.ID)
	})
	
	t.Run("Link ID is unique", func(t *testing.T) {
		link1 := &models.Link{BusinessUnit: "bu01", Network: "mi"}
		link2 := &models.Link{BusinessUnit: "bu01", Network: "mi"}
		
		err1 := linkService.CreateLink(link1)
		err2 := linkService.CreateLink(link2)
		
		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, link1.LinkID, link2.LinkID)
	})
}

func TestLinkService_GetLinkByID(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	linkService := services.NewLinkService(ts.DB, ts.Redis)
	
	// Create test data
	testLink := fixtures.CreateTestLink()
	err := ts.DB.Create(testLink).Error
	require.NoError(t, err)
	
	testTargets := fixtures.CreateTestTargets()
	for _, target := range testTargets {
		err := ts.DB.Create(&target).Error
		require.NoError(t, err)
	}
	
	t.Run("Get existing link", func(t *testing.T) {
		link, err := linkService.GetLinkByID(testLink.LinkID)
		require.NoError(t, err)
		require.NotNil(t, link)
		
		assert.Equal(t, testLink.LinkID, link.LinkID)
		assert.Equal(t, testLink.BusinessUnit, link.BusinessUnit)
		assert.Len(t, link.Targets, 2)
	})
	
	t.Run("Get non-existing link", func(t *testing.T) {
		link, err := linkService.GetLinkByID("nonexistent")
		require.NoError(t, err)
		assert.Nil(t, link)
	})
}

func TestLinkService_SelectTarget(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	linkService := services.NewLinkService(ts.DB, ts.Redis)
	
	// Create test link with targets
	link := fixtures.CreateTestLink()
	link.Targets = fixtures.CreateTestTargets()
	
	t.Run("Select target successfully", func(t *testing.T) {
		target, err := linkService.SelectTarget(link, "192.168.1.1", "US")
		require.NoError(t, err)
		require.NotNil(t, target)
		
		// Should select target 1 (US country)
		assert.Equal(t, "https://target1.example.com", target.URL)
	})
	
	t.Run("Select target for different country", func(t *testing.T) {
		target, err := linkService.SelectTarget(link, "192.168.1.2", "UK")
		require.NoError(t, err)
		require.NotNil(t, target)
		
		// Should select target 2 (UK country)
		assert.Equal(t, "https://target2.example.com", target.URL)
	})
	
	t.Run("No targets for unsupported country", func(t *testing.T) {
		target, err := linkService.SelectTarget(link, "192.168.1.3", "JP")
		assert.Error(t, err)
		assert.Nil(t, target)
		assert.Contains(t, err.Error(), "no targets available for this country")
	})
	
	t.Run("No active targets", func(t *testing.T) {
		link := fixtures.CreateTestLink()
		link.Targets = fixtures.CreateTestTargets()
		
		// Deactivate all targets
		for i := range link.Targets {
			link.Targets[i].IsActive = false
		}
		
		target, err := linkService.SelectTarget(link, "192.168.1.4", "US")
		assert.Error(t, err)
		assert.Nil(t, target)
	})
}

func TestLinkService_ProcessParameters(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	linkService := services.NewLinkService(ts.DB, ts.Redis)
	
	target := &models.Target{
		ParamMapping: `{"kw":"q","src":"source"}`,
		StaticParams: `{"ref":"test","campaign":"summer"}`,
	}
	
	originalParams := map[string]string{
		"kw":    "golang",
		"src":   "google",
		"extra": "value",
	}
	
	result, err := linkService.ProcessParameters(target, originalParams)
	require.NoError(t, err)
	
	// Check parameter mapping
	assert.Equal(t, "golang", result["q"])
	assert.Equal(t, "google", result["source"])
	assert.Equal(t, "value", result["extra"])
	
	// Check static parameters
	assert.Equal(t, "test", result["ref"])
	assert.Equal(t, "summer", result["campaign"])
	
	// Original mapped parameters should be removed
	_, exists := result["kw"]
	assert.False(t, exists)
	_, exists = result["src"]
	assert.False(t, exists)
}

func TestLinkService_IncrementHits(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	linkService := services.NewLinkService(ts.DB, ts.Redis)
	
	// Create test data
	link := fixtures.CreateTestLink()
	err := ts.DB.Create(link).Error
	require.NoError(t, err)
	
	target := fixtures.CreateTestTargets()[0]
	target.LinkID = link.ID
	err = ts.DB.Create(&target).Error
	require.NoError(t, err)
	
	initialLinkHits := link.CurrentHits
	initialTargetHits := target.CurrentHits
	
	err = linkService.IncrementHits(link.ID, target.ID)
	require.NoError(t, err)
	
	// Verify hits were incremented
	var updatedLink models.Link
	err = ts.DB.First(&updatedLink, link.ID).Error
	require.NoError(t, err)
	assert.Equal(t, initialLinkHits+1, updatedLink.CurrentHits)
	
	var updatedTarget models.Target
	err = ts.DB.First(&updatedTarget, target.ID).Error
	require.NoError(t, err)
	assert.Equal(t, initialTargetHits+1, updatedTarget.CurrentHits)
}