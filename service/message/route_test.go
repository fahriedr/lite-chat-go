package message

import (
	"bytes"
	"context"
	"encoding/json"
	"lite-chat-go/internal/testutils"
	"lite-chat-go/models"
	"lite-chat-go/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMessageService_GetMessage(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		messageService := NewMessageService(testDB.MsgCol, testDB.ConvCol, testDB.UserCol)

		// Create test users
		user1, _ := testDB.CreateTestUser("user1@example.com", "user1", "User One")
		user2, _ := testDB.CreateTestUser("user2@example.com", "user2", "User Two")

		// Create test messages
		message1, _ := testDB.CreateTestMessage(user1.ID, user2.ID, "Hello from user1")
		message2, _ := testDB.CreateTestMessage(user2.ID, user1.ID, "Hello back from user2")
		message3, _ := testDB.CreateTestMessage(user1.ID, user2.ID, "Another message from user1")

		// Create conversation with messages
		testDB.CreateTestConversation(
			[]primitive.ObjectID{user1.ID, user2.ID},
			[]primitive.ObjectID{message1.ID, message2.ID, message3.ID},
		)

		t.Run("Get messages between two users", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/list/{receiver_id}", messageService.getMessage).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/list/"+user2.ID.Hex(), nil)
			
			// Add user context
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Success", response.Message)
			
			messages := response.Data.([]interface{})
			assert.True(t, len(messages) > 0) // Should have messages
			
			// Verify message structure
			msg := messages[0].(map[string]interface{})
			assert.Contains(t, msg, "_id")
			assert.Contains(t, msg, "senderId")
			assert.Contains(t, msg, "receiverId")
			assert.Contains(t, msg, "message")
			assert.Contains(t, msg, "isRead")
			assert.Contains(t, msg, "createdAt")
		})

		t.Run("Get messages with invalid receiver ID", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/list/{receiver_id}", messageService.getMessage).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/list/invalid-id", nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "User ID not found")
		})

		t.Run("Get messages with invalid user ID in context", func(t *testing.T) {
			router := mux.NewRouter()
			router.HandleFunc("/list/{receiver_id}", messageService.getMessage).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/list/"+user2.ID.Hex(), nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, "invalid-id")
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})

		t.Run("Get messages with no conversation", func(t *testing.T) {
			// Create a third user with no conversation with user1
			user3, _ := testDB.CreateTestUser("user3@example.com", "user3", "User Three")

			router := mux.NewRouter()
			router.HandleFunc("/list/{receiver_id}", messageService.getMessage).Methods(http.MethodGet)

			req := httptest.NewRequest(http.MethodGet, "/list/"+user3.ID.Hex(), nil)
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			
			messages := response.Data.([]interface{})
			assert.Len(t, messages, 0) // No messages
		})
	})
}

func TestMessageService_SendMessage(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		messageService := NewMessageService(testDB.MsgCol, testDB.ConvCol, testDB.UserCol)

		// Create test users
		user1, _ := testDB.CreateTestUser("sender@example.com", "sender", "Sender User")
		user2, _ := testDB.CreateTestUser("receiver@example.com", "receiver", "Receiver User")

		t.Run("Send message to existing user", func(t *testing.T) {
			payload := models.MessagePayload{
				UserId:  user2.ID.Hex(),
				Message: "Hello, this is a test message",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(body))
			
			// Add user context
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.sendMessage(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Success", response.Message)
			assert.Equal(t, 200, response.Status)
			
			// Verify message data
			messageData := response.Data.(map[string]interface{})
			assert.Contains(t, messageData, "_id")
			assert.Equal(t, payload.Message, messageData["message"])
			assert.Equal(t, user1.ID.Hex(), messageData["senderId"])
			assert.Equal(t, user2.ID.Hex(), messageData["receiverId"])
			assert.False(t, messageData["isRead"].(bool))
		})

		t.Run("Send message to non-existent user", func(t *testing.T) {
			nonExistentID := primitive.NewObjectID()
			payload := models.MessagePayload{
				UserId:  nonExistentID.Hex(),
				Message: "Message to non-existent user",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.sendMessage(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Equal(t, "User not found", response["message"])
		})

		t.Run("Send message to self", func(t *testing.T) {
			payload := models.MessagePayload{
				UserId:  user1.ID.Hex(), // Same as sender
				Message: "Message to myself",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.sendMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "not valid")
		})

		t.Run("Send message with invalid JSON", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer([]byte("invalid json")))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.sendMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Send message with missing required fields", func(t *testing.T) {
			payload := models.MessagePayload{
				UserId: user2.ID.Hex(),
				// Missing Message field
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.sendMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Send message with invalid user ID format", func(t *testing.T) {
			payload := models.MessagePayload{
				UserId:  "invalid-id-format",
				Message: "Test message",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.sendMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Send message with invalid sender ID in context", func(t *testing.T) {
			payload := models.MessagePayload{
				UserId:  user2.ID.Hex(),
				Message: "Test message",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/send", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, "invalid-id")
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.sendMessage(w, req)

			assert.Equal(t, http.StatusInternalServerError, w.Code)
		})
	})
}

func TestMessageService_UpdateStatusMessage(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		messageService := NewMessageService(testDB.MsgCol, testDB.ConvCol, testDB.UserCol)

		// Create test users
		user1, _ := testDB.CreateTestUser("sender@example.com", "sender", "Sender User")
		user2, _ := testDB.CreateTestUser("receiver@example.com", "receiver", "Receiver User")

		// Create test message
		testMessage, _ := testDB.CreateTestMessage(user1.ID, user2.ID, "Test message for status update")

		t.Run("Update message status successfully", func(t *testing.T) {
			payload := models.UpdateMessagePayload{
				MessageID: testMessage.ID.Hex(),
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/update-status", bytes.NewBuffer(body))
			
			// Add receiver context (only receiver can mark as read)
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user2.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.updateStatusMessage(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response types.CustomSuccessResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.True(t, response.Success)
			assert.Equal(t, "Success Update message status", response.Message)
			assert.Equal(t, http.StatusOK, response.Status)

			// Verify message was updated in database
			var updatedMessage models.Message
			err = testDB.MsgCol.FindOne(context.Background(), bson.M{"_id": testMessage.ID}).Decode(&updatedMessage)
			assert.NoError(t, err)
			assert.True(t, updatedMessage.IsRead)
		})

		t.Run("Update status with non-existent message ID", func(t *testing.T) {
			nonExistentID := primitive.NewObjectID()
			payload := models.UpdateMessagePayload{
				MessageID: nonExistentID.Hex(),
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/update-status", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user2.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.updateStatusMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
		})

		t.Run("Update status by non-receiver user", func(t *testing.T) {
			// Create another message
			anotherMessage, _ := testDB.CreateTestMessage(user1.ID, user2.ID, "Another test message")

			payload := models.UpdateMessagePayload{
				MessageID: anotherMessage.ID.Hex(),
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/update-status", bytes.NewBuffer(body))
			
			// Try to update status as sender (should fail)
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user1.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.updateStatusMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.False(t, response["success"].(bool))
			assert.Contains(t, response["message"], "Error processing data")
		})

		t.Run("Update status with invalid message ID format", func(t *testing.T) {
			payload := models.UpdateMessagePayload{
				MessageID: "invalid-id-format",
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/update-status", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user2.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.updateStatusMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Update status with invalid JSON", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/update-status", bytes.NewBuffer([]byte("invalid json")))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user2.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.updateStatusMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Update status with missing message ID", func(t *testing.T) {
			payload := models.UpdateMessagePayload{
				// Missing MessageID
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/update-status", bytes.NewBuffer(body))
			
			ctx := context.WithValue(req.Context(), types.ContextKeyUserID, user2.ID.Hex())
			req = req.WithContext(ctx)
			
			w := httptest.NewRecorder()

			messageService.updateStatusMessage(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	})
}

func TestNewMessageService(t *testing.T) {
	testutils.RunTestWithDB(t, func(testDB *testutils.TestDB) {
		t.Run("Create new message service", func(t *testing.T) {
			service := NewMessageService(testDB.MsgCol, testDB.ConvCol, testDB.UserCol)
			
			assert.NotNil(t, service)
			assert.Equal(t, testDB.MsgCol, service.messageCollection)
			assert.Equal(t, testDB.ConvCol, service.conversationCollection)
			assert.Equal(t, testDB.UserCol, service.userCollection)
		})

		t.Run("Create service with nil collections", func(t *testing.T) {
			service := NewMessageService(nil, nil, nil)
			
			assert.NotNil(t, service)
			assert.Nil(t, service.messageCollection)
			assert.Nil(t, service.conversationCollection)
			assert.Nil(t, service.userCollection)
		})
	})
}