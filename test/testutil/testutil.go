package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/raoxb/smart_redirect/internal/api"
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/pkg/auth"
)

type TestSuite struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Router *gin.Engine
	JWT    *auth.JWTManager
}

func SetupTestSuite(t *testing.T) *TestSuite {
	gin.SetMode(gin.TestMode)
	
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)
	
	err = db.AutoMigrate(
		&models.User{},
		&models.Link{},
		&models.Target{},
		&models.LinkPermission{},
		&models.AccessLog{},
		&api.LinkTemplate{},
	)
	assert.NoError(t, err)
	
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use different DB for tests
	})
	
	router := gin.New()
	jwtManager := auth.NewJWTManager("test-secret", 24)
	
	return &TestSuite{
		DB:     db,
		Redis:  redisClient,
		Router: router,
		JWT:    jwtManager,
	}
}

func (ts *TestSuite) TearDown() {
	ts.Redis.FlushDB(context.Background())
	ts.Redis.Close()
}

func (ts *TestSuite) CreateTestToken(userID uint, username, role string) string {
	token, _ := ts.JWT.GenerateToken(userID, username, role)
	return token
}

func MakeRequest(t *testing.T, router *gin.Engine, method, url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		assert.NoError(t, err)
		bodyReader = bytes.NewReader(jsonBody)
	} else {
		bodyReader = bytes.NewReader([]byte{})
	}
	
	req, err := http.NewRequest(method, url, bodyReader)
	assert.NoError(t, err)
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	return w
}

func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody interface{}) {
	assert.Equal(t, expectedStatus, w.Code)
	
	if expectedBody != nil {
		var actualBody interface{}
		err := json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.NoError(t, err)
		
		expectedJSON, _ := json.Marshal(expectedBody)
		var expectedBodyParsed interface{}
		json.Unmarshal(expectedJSON, &expectedBodyParsed)
		
		assert.Equal(t, expectedBodyParsed, actualBody)
	}
}

func (ts *TestSuite) SeedTestData(t *testing.T) {
	// Create test users
	users := []models.User{
		{Username: "testuser", Email: "test@example.com", Password: "hashedpassword", Role: "user", IsActive: true},
		{Username: "admin", Email: "admin@example.com", Password: "hashedpassword", Role: "admin", IsActive: true},
	}
	
	for _, user := range users {
		err := ts.DB.Create(&user).Error
		assert.NoError(t, err)
	}
	
	// Create test links
	link := models.Link{
		LinkID:       "test123",
		BusinessUnit: "bu01",
		Network:      "mi",
		TotalCap:     1000,
		CurrentHits:  0,
		BackupURL:    "https://backup.example.com",
		IsActive:     true,
	}
	err := ts.DB.Create(&link).Error
	assert.NoError(t, err)
	
	// Create test targets
	targets := []models.Target{
		{
			LinkID:       link.ID,
			URL:          "https://target1.example.com",
			Weight:       70,
			Cap:          500,
			CurrentHits:  0,
			Countries:    `["US","CA"]`,
			ParamMapping: `{"kw":"q"}`,
			StaticParams: `{"ref":"test"}`,
			IsActive:     true,
		},
		{
			LinkID:       link.ID,
			URL:          "https://target2.example.com",
			Weight:       30,
			Cap:          300,
			CurrentHits:  0,
			Countries:    `["UK","DE"]`,
			IsActive:     true,
		},
	}
	
	for _, target := range targets {
		err := ts.DB.Create(&target).Error
		assert.NoError(t, err)
	}
}

func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}