package unit

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	
	"github.com/raoxb/smart_redirect/internal/models"
	"github.com/raoxb/smart_redirect/internal/services"
	"github.com/raoxb/smart_redirect/test/testutil"
)

func TestUserService_CreateUser(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	userService := services.NewUserService(ts.DB)
	
	t.Run("Create user successfully", func(t *testing.T) {
		user := &models.User{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "password123",
			Role:     "user",
		}
		
		err := userService.CreateUser(user)
		require.NoError(t, err)
		
		assert.NotZero(t, user.ID)
		assert.NotEqual(t, "password123", user.Password) // Should be hashed
		
		// Verify password was hashed correctly
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123"))
		assert.NoError(t, err)
	})
	
	t.Run("Create user with duplicate username", func(t *testing.T) {
		user1 := &models.User{
			Username: "duplicate",
			Email:    "user1@example.com",
			Password: "password123",
			Role:     "user",
		}
		
		user2 := &models.User{
			Username: "duplicate",
			Email:    "user2@example.com",
			Password: "password123",
			Role:     "user",
		}
		
		err1 := userService.CreateUser(user1)
		require.NoError(t, err1)
		
		err2 := userService.CreateUser(user2)
		assert.Error(t, err2) // Should fail due to unique constraint
	})
}

func TestUserService_GetUserByUsername(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	userService := services.NewUserService(ts.DB)
	
	// Create test user
	testUser := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     "user",
	}
	err := ts.DB.Create(testUser).Error
	require.NoError(t, err)
	
	t.Run("Get existing user", func(t *testing.T) {
		user, err := userService.GetUserByUsername("testuser")
		require.NoError(t, err)
		require.NotNil(t, user)
		
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})
	
	t.Run("Get non-existing user", func(t *testing.T) {
		user, err := userService.GetUserByUsername("nonexistent")
		require.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestUserService_VerifyPassword(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	userService := services.NewUserService(ts.DB)
	
	// Create user with known password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	user := &models.User{
		Username: "testuser",
		Password: string(hashedPassword),
	}
	
	t.Run("Correct password", func(t *testing.T) {
		valid := userService.VerifyPassword(user, "correctpassword")
		assert.True(t, valid)
	})
	
	t.Run("Incorrect password", func(t *testing.T) {
		valid := userService.VerifyPassword(user, "wrongpassword")
		assert.False(t, valid)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	userService := services.NewUserService(ts.DB)
	
	// Create test user
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "oldpassword",
		Role:     "user",
	}
	err := userService.CreateUser(user)
	require.NoError(t, err)
	
	t.Run("Update user fields", func(t *testing.T) {
		user.Email = "newemail@example.com"
		user.Role = "admin"
		user.Password = "newpassword"
		
		err := userService.UpdateUser(user)
		require.NoError(t, err)
		
		// Fetch updated user
		updatedUser, err := userService.GetUserByID(user.ID)
		require.NoError(t, err)
		
		assert.Equal(t, "newemail@example.com", updatedUser.Email)
		assert.Equal(t, "admin", updatedUser.Role)
		
		// Verify new password
		valid := userService.VerifyPassword(updatedUser, "newpassword")
		assert.True(t, valid)
	})
}

func TestUserService_ListUsers(t *testing.T) {
	ts := testutil.SetupTestSuite(t)
	defer ts.TearDown()
	
	userService := services.NewUserService(ts.DB)
	
	// Create multiple test users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", Password: "password", Role: "user"},
		{Username: "user2", Email: "user2@example.com", Password: "password", Role: "user"},
		{Username: "user3", Email: "user3@example.com", Password: "password", Role: "admin"},
	}
	
	for _, user := range users {
		err := userService.CreateUser(user)
		require.NoError(t, err)
	}
	
	t.Run("List all users", func(t *testing.T) {
		userList, total, err := userService.ListUsers(0, 10)
		require.NoError(t, err)
		
		assert.Equal(t, int64(3), total)
		assert.Len(t, userList, 3)
	})
	
	t.Run("List with pagination", func(t *testing.T) {
		userList, total, err := userService.ListUsers(0, 2)
		require.NoError(t, err)
		
		assert.Equal(t, int64(3), total)
		assert.Len(t, userList, 2)
		
		// Get second page
		userList2, total2, err := userService.ListUsers(2, 2)
		require.NoError(t, err)
		
		assert.Equal(t, int64(3), total2)
		assert.Len(t, userList2, 1)
	})
}