package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
	
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/raoxb/smart_redirect/internal/api"
	"github.com/raoxb/smart_redirect/internal/middleware"
	"github.com/raoxb/smart_redirect/test/testutil"
)

func TestRedirectHandler_Integration(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	// Seed test data
	ts.SeedTestData(t)
	
	// Setup routes
	redirectHandler := api.NewRedirectHandler(ts.DB, ts.Redis)
	ts.Router.GET("/v1/:bu/:link_id", 
		middleware.RateLimitMiddleware(ts.Redis, 10, time.Hour),
		redirectHandler.HandleRedirect)
	
	t.Run("Successful redirect", func(t *testing.T) {
		w := testutil.MakeRequest(t, ts.Router, "GET", "/v1/bu01/test123?network=mi", nil, nil)
		
		assert.Equal(t, http.StatusFound, w.Code)
		
		location := w.Header().Get("Location")
		assert.NotEmpty(t, location)
		assert.Contains(t, location, "target1.example.com") // Should redirect to US target
	})
	
	t.Run("Link not found", func(t *testing.T) {
		w := testutil.MakeRequest(t, ts.Router, "GET", "/v1/bu01/nonexistent?network=mi", nil, nil)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	
	t.Run("Wrong business unit", func(t *testing.T) {
		w := testutil.MakeRequest(t, ts.Router, "GET", "/v1/bu99/test123?network=mi", nil, nil)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	
	t.Run("Rate limiting", func(t *testing.T) {
		// Make multiple requests to trigger rate limit
		for i := 0; i < 12; i++ {
			w := testutil.MakeRequest(t, ts.Router, "GET", "/v1/bu01/test123?network=mi", nil, 
				map[string]string{"X-Real-IP": "192.168.1.100"})
			
			if i < 10 {
				assert.Equal(t, http.StatusFound, w.Code)
			} else {
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			}
		}
	})
	
	t.Run("Parameter processing", func(t *testing.T) {
		w := testutil.MakeRequest(t, ts.Router, "GET", "/v1/bu01/test123?network=mi&kw=golang&extra=value", nil, nil)
		
		assert.Equal(t, http.StatusFound, w.Code)
		
		location := w.Header().Get("Location")
		assert.Contains(t, location, "q=golang") // kw should be mapped to q
		assert.Contains(t, location, "ref=test") // static param should be added
		assert.Contains(t, location, "extra=value") // original param should be preserved
	})
}

func TestAuthHandler_Integration(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	// Setup routes
	authHandler := api.NewAuthHandler(ts.DB, ts.JWT)
	ts.Router.POST("/api/v1/auth/login", authHandler.Login)
	ts.Router.POST("/api/v1/auth/register", authHandler.Register)
	ts.Router.GET("/api/v1/auth/profile", 
		middleware.AuthMiddleware(ts.JWT), 
		authHandler.GetProfile)
	
	t.Run("User registration", func(t *testing.T) {
		registerData := map[string]interface{}{
			"username": "newuser",
			"email":    "newuser@example.com",
			"password": "password123",
		}
		
		w := testutil.MakeRequest(t, ts.Router, "POST", "/api/v1/auth/register", registerData, nil)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := testutil.UnmarshalJSON(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "newuser", response["username"])
		assert.Equal(t, "newuser@example.com", response["email"])
		assert.Equal(t, "user", response["role"])
	})
	
	t.Run("User login", func(t *testing.T) {
		// First register a user
		registerData := map[string]interface{}{
			"username": "loginuser",
			"email":    "login@example.com",
			"password": "password123",
		}
		
		w := testutil.MakeRequest(t, ts.Router, "POST", "/api/v1/auth/register", registerData, nil)
		require.Equal(t, http.StatusCreated, w.Code)
		
		// Then login
		loginData := map[string]interface{}{
			"username": "loginuser",
			"password": "password123",
		}
		
		w = testutil.MakeRequest(t, ts.Router, "POST", "/api/v1/auth/login", loginData, nil)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := testutil.UnmarshalJSON(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.NotEmpty(t, response["token"])
		assert.Equal(t, "loginuser", response["username"])
	})
	
	t.Run("Invalid login", func(t *testing.T) {
		loginData := map[string]interface{}{
			"username": "nonexistent",
			"password": "wrongpassword",
		}
		
		w := testutil.MakeRequest(t, ts.Router, "POST", "/api/v1/auth/login", loginData, nil)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	
	t.Run("Get profile with valid token", func(t *testing.T) {
		token := ts.CreateTestToken(1, "testuser", "user")
		headers := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		
		w := testutil.MakeRequest(t, ts.Router, "GET", "/api/v1/auth/profile", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
	})
	
	t.Run("Get profile without token", func(t *testing.T) {
		w := testutil.MakeRequest(t, ts.Router, "GET", "/api/v1/auth/profile", nil, nil)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestLinkHandler_Integration(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	// Setup routes
	linkHandler := api.NewLinkHandler(ts.DB, ts.Redis)
	authGroup := ts.Router.Group("/api/v1")
	authGroup.Use(middleware.AuthMiddleware(ts.JWT))
	{
		authGroup.POST("/links", linkHandler.CreateLink)
		authGroup.GET("/links", linkHandler.ListLinks)
		authGroup.GET("/links/:link_id", linkHandler.GetLink)
		authGroup.POST("/links/:link_id/targets", linkHandler.CreateTarget)
	}
	
	userToken := ts.CreateTestToken(1, "testuser", "user")
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", userToken),
	}
	
	t.Run("Create link", func(t *testing.T) {
		linkData := map[string]interface{}{
			"business_unit": "bu01",
			"network":       "mi",
			"total_cap":     1000,
			"backup_url":    "https://backup.example.com",
		}
		
		w := testutil.MakeRequest(t, ts.Router, "POST", "/api/v1/links", linkData, headers)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := testutil.UnmarshalJSON(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "bu01", response["business_unit"])
		assert.Equal(t, "mi", response["network"])
		assert.NotEmpty(t, response["link_id"])
	})
	
	t.Run("List links", func(t *testing.T) {
		w := testutil.MakeRequest(t, ts.Router, "GET", "/api/v1/links", nil, headers)
		
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := testutil.UnmarshalJSON(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Contains(t, response, "data")
		assert.Contains(t, response, "total")
	})
	
	t.Run("Create target", func(t *testing.T) {
		// First seed a link
		ts.SeedTestData(t)
		
		targetData := map[string]interface{}{
			"url":     "https://new-target.example.com",
			"weight":  50,
			"cap":     200,
			"countries": []string{"US", "CA"},
			"param_mapping": map[string]string{"kw": "query"},
			"static_params": map[string]string{"source": "test"},
		}
		
		w := testutil.MakeRequest(t, ts.Router, "POST", "/api/v1/links/test123/targets", targetData, headers)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := testutil.UnmarshalJSON(w.Body.Bytes(), &response)
		require.NoError(t, err)
		
		assert.Equal(t, "https://new-target.example.com", response["url"])
		assert.Equal(t, float64(50), response["weight"])
	})
}