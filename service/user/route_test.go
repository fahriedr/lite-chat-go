package user

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"lite-chat-go/internal/testutils"
	"lite-chat-go/models"
	"lite-chat-go/types"
	"lite-chat-go/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUserService_Register(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		userService := NewUserService(testDB.UserCol)

		t.Run("Valid registration", func(t *testing.T) {
			payload := models.UserRegisterPayload{
				Fullname:        "John Doe",
				Email:           "john@example.com",
				Username:        "johndoe",
				Password:        "password123",
				ConfirmPassword: "password123",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleRegister(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "Login successful", response["message"])
			assert.Contains(t, response, "token")
			assert.Contains(t, response, "user")

			// Verify user data
			userData := response["user"].(map[string]interface{})
			assert.Equal(t, payload.Fullname, userData["fullname"])
			assert.Equal(t, payload.Email, userData["email"])
			assert.Equal(t, payload.Username, userData["username"])
		})

		t.Run("Duplicate email registration", func(t *testing.T) {
			// Create a user first
			testDB.CreateTestUser("existing@example.com", "existinguser", "Existing User")

			payload := models.UserRegisterPayload{
				Fullname:        "New User",
				Email:           "existing@example.com", // Same email
				Username:        "newuser",
				Password:        "password123",
				ConfirmPassword: "password123",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleRegister(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "already exists")
		})

		t.Run("Duplicate username registration", func(t *testing.T) {
			// Create a user first
			testDB.CreateTestUser("user1@example.com", "duplicateuser", "User One")

			payload := models.UserRegisterPayload{
				Fullname:        "User Two",
				Email:           "user2@example.com",
				Username:        "duplicateuser", // Same username
				Password:        "password123",
				ConfirmPassword: "password123",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleRegister(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "already taken")
		})

		t.Run("Invalid email format", func(t *testing.T) {
			payload := models.UserRegisterPayload{
				Fullname:        "John Doe",
				Email:           "invalid-email", // Invalid format
				Username:        "johndoe",
				Password:        "password123",
				ConfirmPassword: "password123",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleRegister(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Missing required fields", func(t *testing.T) {
			payload := models.UserRegisterPayload{
				Email:    "test@example.com",
				Username: "testuser",
				// Missing fullname and password
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleRegister(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}

func TestUserService_Login(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		userService := NewUserService(testDB.UserCol)

		// Create test user
		testUser, _ := testDB.CreateTestUser("login@example.com", "loginuser", "Login User")

		t.Run("Valid login", func(t *testing.T) {
			payload := models.UserLoginPayload{
				Email:    "login@example.com",
				Password: "testpassword", // This matches the password used in CreateTestUser
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleLogin(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "Login", response["success"])
			assert.Contains(t, response, "token")
			assert.Contains(t, response, "user")

			// Verify user data
			userData := response["user"].(map[string]interface{})
			assert.Equal(t, testUser.Email, userData["email"])
			assert.Equal(t, testUser.Username, userData["username"])
		})

		t.Run("Invalid email", func(t *testing.T) {
			payload := models.UserLoginPayload{
				Email:    "nonexistent@example.com",
				Password: "testpassword",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleLogin(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "incorrect")
		})

		t.Run("Invalid password", func(t *testing.T) {
			payload := models.UserLoginPayload{
				Email:    "login@example.com",
				Password: "wrongpassword",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleLogin(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "incorrect")
		})

		t.Run("OAuth user without password", func(t *testing.T) {
			// Create OAuth user (no password)
			oauthUser := &models.User{
				ID:            primitive.NewObjectID(),
				Fullname:      "OAuth User",
				Username:      "oauthuser",
				Email:         "oauth@example.com",
				Password:      nil, // No password
				Provider:      (*models.AuthProvider)(&[]models.AuthProvider{models.ProviderGoogle}[0]),
				IsActive:      true,
				EmailVerified: true,
			}
			testDB.UserCol.InsertOne(context.Background(), oauthUser)

			payload := models.UserLoginPayload{
				Email:    "oauth@example.com",
				Password: "anypassword",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			userService.handleLogin(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "Sign in with")
		})

		t.Run("Invalid JSON payload", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("invalid json")))
			w := httptest.NewRecorder()

			userService.handleLogin(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}

func TestUserService_Profile(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		userService := NewUserService(testDB.UserCol)

		// Create test user
		testUser, _ := testDB.CreateTestUser("profile@example.com", "profileuser", "Profile User")

		t.Run("Valid profile request", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			
			// Add user context (simulating JWT middleware)
			ctx := context.WithValue(req.Context(), types.ContextKeyEmail, testUser.Email)
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			userService.profile(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "Login", response["message"])
			
			userData := response["data"].(map[string]interface{})
			assert.Equal(t, testUser.Email, userData["email"])
			assert.Equal(t, testUser.Username, userData["username"])
			assert.Equal(t, testUser.Fullname, userData["fullname"])
		})

		t.Run("User not found", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			
			// Add non-existent user context
			ctx := context.WithValue(req.Context(), types.ContextKeyEmail, "nonexistent@example.com")
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			userService.profile(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
		})
	})
}

func TestUserService_Search(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		userService := NewUserService(testDB.UserCol)

		// Create test users
		testUser1, _ := testDB.CreateTestUser("user1@example.com", "user1", "User One")
		testUser2, _ := testDB.CreateTestUser("user2@example.com", "user2", "User Two")
		testUser3, _ := testDB.CreateTestUser("searcher@example.com", "searcher", "Searcher User")

		t.Run("Search by username", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/search/{query}", userService.handleSearch).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/search/user1", nil)
			
			// Add searcher context (exclude this user from results)
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, testUser3.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Success", response.Message)
			
			users := response.Data.([]interface{})
			assert.Len(t, users, 1)
			
			foundUser := users[0].(map[string]interface{})
			assert.Equal(t, testUser1.Username, foundUser["username"])
		})

		t.Run("Search by email", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/search/{query}", userService.handleSearch).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/search/user2@example.com", nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, testUser3.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			
			users := response.Data.([]interface{})
			assert.Len(t, users, 1)
			
			foundUser := users[0].(map[string]interface{})
			assert.Equal(t, testUser2.Email, foundUser["email"])
		})

		t.Run("Search with partial match", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/search/{query}", userService.handleSearch).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/search/user", nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, testUser3.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			
			users := response.Data.([]interface{})
			assert.Len(t, users, 2) // Should find user1 and user2, but not searcher
		})

		t.Run("Search excludes current user", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/search/{query}", userService.handleSearch).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/search/searcher", nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, testUser3.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			
			users := response.Data.([]interface{})
			assert.Len(t, users, 0) // Should not find the current user
		})

		t.Run("No results found", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/search/{query}", userService.handleSearch).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/search/nonexistent", nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, testUser3.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			
			users := response.Data.([]interface{})
			assert.Len(t, users, 0)
		})

		t.Run("Invalid user ID in context", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/search/{query}", userService.handleSearch).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/search/test", nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, "invalid-id")
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}