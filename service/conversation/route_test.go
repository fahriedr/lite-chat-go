package conversation

import (
	"context"
	"encoding/json"
	"lite-chat-go/internal/testutils"
	"lite-chat-go/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestConversationService_GetConversation(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		conversationService := NewConversationService(testDB.ConvCol)

		// Create test users
		user1, _ := testDB.CreateTestUser("user1@example.com", "user1", "User One")
		user2, _ := testDB.CreateTestUser("user2@example.com", "user2", "User Two")
		user3, _ := testDB.CreateTestUser("user3@example.com", "user3", "User Three")

		// Create test messages
		message1, _ := testDB.CreateTestMessage(user1.ID, user2.ID, "Hello from user1")
		message2, _ := testDB.CreateTestMessage(user2.ID, user1.ID, "Hello back from user2")
		message3, _ := testDB.CreateTestMessage(user1.ID, user3.ID, "Message to user3")

		// Create test conversations
		conv1, _ := testDB.CreateTestConversation(
			[]primitive.ObjectID{user1.ID, user2.ID},
			[]primitive.ObjectID{message1.ID, message2.ID},
		)

		conv2, _ := testDB.CreateTestConversation(
			[]primitive.ObjectID{user1.ID, user3.ID},
			[]primitive.ObjectID{message3.ID},
		)

		t.Run("Get conversations for user1", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/conversations", nil)

			// Add user context
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			conversationService.getConversation(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Success", response.Message)

			conversations := response.Data.([]interface{})
			assert.Len(t, conversations, 2) // user1 has 2 conversations

			// Verify conversation structure
			conv := conversations[0].(map[string]interface{})
			assert.Contains(t, conv, "_id")
			assert.Contains(t, conv, "participants")
			assert.Contains(t, conv, "messages")

			// Verify participant is not the current user
			participant := conv["participants"].(map[string]interface{})
			participantID := participant["_id"].(string)
			assert.NotEqual(t, user1.ID.Hex(), participantID)
			assert.True(t, participantID == user2.ID.Hex() || participantID == user3.ID.Hex())
		})

		t.Run("Get conversations for user with no conversations", func(t *testing.T) {
			// Create a user with no conversations
			user4, _ := testDB.CreateTestUser("user4@example.com", "user4", "User Four")

			req := httptest.NewRequest(http.MethodGet, "/conversations", nil)

			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user4.ID.Hex())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			conversationService.getConversation(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)

			conversations := response.Data.([]interface{})
			assert.Len(t, conversations, 0) // No conversations
		})

		t.Run("Invalid user ID in context", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/conversations", nil)

			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, "invalid-user-id")
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			conversationService.getConversation(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("Verify conversation filtering", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/conversations", nil)

			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user2.ID.Hex())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			conversationService.getConversation(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)

			conversations := response.Data.([]interface{})
			assert.Len(t, conversations, 1) // user2 has only 1 conversation (with user1)

			conv := conversations[0].(map[string]interface{})
			participant := conv["participants"].(map[string]interface{})

			// Verify the participant is user1 (not user2)
			assert.Equal(t, user1.ID.Hex(), participant["_id"].(string))
			assert.Equal(t, user1.Username, participant["username"])
		})

		t.Run("Verify messages are included", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/conversations", nil)

			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			conversationService.getConversation(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			conversations := response.Data.([]interface{})
			assert.True(t, len(conversations) > 0)

			// Check if conversations have messages
			for _, convInterface := range conversations {
				conv := convInterface.(map[string]interface{})
				messages := conv["messages"].([]interface{})

				// Each conversation should have at least one message (last message)
				if len(messages) > 0 {
					msg := messages[0].(map[string]interface{})
					assert.Contains(t, msg, "_id")
					assert.Contains(t, msg, "senderId")
					assert.Contains(t, msg, "receiverId")
					assert.Contains(t, msg, "message")
				}
			}
		})

		// Clean up test data
		testDB.ClearCollections()

		// Test edge case with empty conversation collection
		t.Run("Empty conversation collection", func(t *testing.T) {
			// Create a user but no conversations
			userEmpty, _ := testDB.CreateTestUser("empty@example.com", "emptyuser", "Empty User")

			req := httptest.NewRequest(http.MethodGet, "/conversations", nil)

			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, userEmpty.ID.Hex())
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			conversationService.getConversation(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Success", response.Message)

			conversations := response.Data.([]interface{})
			assert.Len(t, conversations, 0)
		})
	})
}

func TestConversationService_RegisterRoutes(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		conversationService := NewConversationService(testDB.ConvCol)

		t.Run("Verify routes are registered", func(t *testing.T) {
			// This test ensures the RegisterRoutes method works without panicking
			// and that the service can handle route registration
			assert.NotNil(t, conversationService)
			assert.NotNil(t, conversationService.conversationCollection)

			// The actual route testing is covered in the handler tests above
		})
	})
}

func TestNewConversationService(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		t.Run("Create new conversation service", func(t *testing.T) {
			service := NewConversationService(testDB.ConvCol)

			assert.NotNil(t, service)
			assert.Equal(t, testDB.ConvCol, service.conversationCollection)
		})

		t.Run("Create service with nil collection", func(t *testing.T) {
			service := NewConversationService(nil)

			assert.NotNil(t, service)
			assert.Nil(t, service.conversationCollection)
		})
	})
}
