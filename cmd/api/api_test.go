package api

import (
	"context"
	"encoding/json"
	"lite-chat-go/internal/testutils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPIServer_HealthCheck(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		apiServer := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb", "8085")

		t.Run("Health check returns OK", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			apiServer.healthCheck(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "Status OK", response["message"])
		})

		t.Run("Health check with different HTTP methods", func(t *testing.T) {
			methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

			for _, method := range methods {
				req := httptest.NewRequest(method, "/health", nil)
				w := httptest.NewRecorder()

				apiServer.healthCheck(w, req)

				// Health check should work with any method (though route config may restrict)
				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "Status OK", response["message"])
			}
		})

		t.Run("Health check response structure", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			apiServer.healthCheck(w, req)

			assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Verify response structure
			assert.Contains(t, response, "message")
			assert.IsType(t, "", response["message"])
		})
	})
}

func TestNewAPIServer(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		
		t.Run("Create new API server with valid parameters", func(t *testing.T) {
			dbName := "testdb"
			port := "8085"

			server := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, dbName, port)

			assert.NotNil(t, server)
			assert.Equal(t, testDB.UserCol, server.userCollection)
			assert.Equal(t, testDB.ConvCol, server.conversationCollection)
			assert.Equal(t, testDB.MsgCol, server.messageCollection)
			assert.Equal(t, dbName, server.dbName)
			assert.Equal(t, port, server.port)
		})

		t.Run("Create API server with nil collections", func(t *testing.T) {
			server := NewAPIServer(nil, nil, nil, "testdb", "8080")

			assert.NotNil(t, server)
			assert.Nil(t, server.userCollection)
			assert.Nil(t, server.conversationCollection)
			assert.Nil(t, server.messageCollection)
		})

		t.Run("Create API server with empty strings", func(t *testing.T) {
			server := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "", "")

			assert.NotNil(t, server)
			assert.Equal(t, "", server.dbName)
			assert.Equal(t, "", server.port)
		})

		t.Run("API server struct validation", func(t *testing.T) {
			server := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb", "8085")

			// Verify all fields are accessible
			assert.NotNil(t, server.userCollection)
			assert.NotNil(t, server.conversationCollection)
			assert.NotNil(t, server.messageCollection)
			assert.NotEmpty(t, server.dbName)
			assert.NotEmpty(t, server.port)
		})
	})
}

func TestAPIServer_Integration(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		
		t.Run("API server components integration", func(t *testing.T) {
			server := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb", "8085")
			
			// Test that server can handle health check
			req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
			w := httptest.NewRecorder()

			server.healthCheck(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "Status OK", response["message"])
		})

		t.Run("Multiple API server instances", func(t *testing.T) {
			server1 := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb1", "8081")
			server2 := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb2", "8082")

			assert.NotEqual(t, server1.dbName, server2.dbName)
			assert.NotEqual(t, server1.port, server2.port)
			
			// Both should be functional
			assert.Equal(t, testDB.UserCol, server1.userCollection)
			assert.Equal(t, testDB.UserCol, server2.userCollection)
		})
	})
}

func TestAPIServer_ErrorHandling(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		
		t.Run("Health check with context cancellation", func(t *testing.T) {
			server := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb", "8085")

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			
			// Create a cancelled context
			ctx, cancel := context.WithCancel(req.Context())
			cancel() // Cancel immediately
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			// Health check should still work as it doesn't depend on context
			server.healthCheck(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("Health check with nil request body", func(t *testing.T) {
			server := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb", "8085")

			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			w := httptest.NewRecorder()

			server.healthCheck(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "Status OK", response["message"])
		})
	})
}

// Benchmark tests for performance validation
func BenchmarkAPIServer_HealthCheck(b *testing.B) {
	testutils.SetupTestEnv()
	testDB, _ := testutils.SetupTestDB()
	defer testDB.Cleanup()

	server := NewAPIServer(testDB.UserCol, testDB.ConvCol, testDB.MsgCol, "testdb", "8085")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		server.healthCheck(w, req)
	}
}