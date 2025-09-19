package testutils

import (
	"context"
	"fmt"
	"lite-chat-go/config"
	"lite-chat-go/models"
	"lite-chat-go/utils"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tryvium-travels/memongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestDB holds the test database setup
type TestDB struct {
	MongoServer *memongo.Server
	Client      *mongo.Client
	Database    *mongo.Database
	UserCol     *mongo.Collection
	ConvCol     *mongo.Collection
	MsgCol      *mongo.Collection
}

// SetupTestDB creates an in-memory MongoDB instance for testing
func SetupTestDB() (*TestDB, error) {
	mongoServer, err := memongo.Start("4.0.5")
	if err != nil {
		return nil, fmt.Errorf("failed to start memongo: %w", err)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoServer.URI()))
	if err != nil {
		mongoServer.Stop()
		return nil, fmt.Errorf("failed to connect to memongo: %w", err)
	}

	db := client.Database("testdb")
	userCol := db.Collection("users")
	convCol := db.Collection("conversations")
	msgCol := db.Collection("messages")

	return &TestDB{
		MongoServer: mongoServer,
		Client:      client,
		Database:    db,
		UserCol:     userCol,
		ConvCol:     convCol,
		MsgCol:      msgCol,
	}, nil
}

// Cleanup closes the test database connections
func (tdb *TestDB) Cleanup() {
	if tdb.Client != nil {
		tdb.Client.Disconnect(context.Background())
	}
	if tdb.MongoServer != nil {
		tdb.MongoServer.Stop()
	}
}

// ClearCollections removes all data from test collections
func (tdb *TestDB) ClearCollections() error {
	ctx := context.Background()
	
	if err := tdb.UserCol.Drop(ctx); err != nil {
		return err
	}
	if err := tdb.ConvCol.Drop(ctx); err != nil {
		return err
	}
	if err := tdb.MsgCol.Drop(ctx); err != nil {
		return err
	}
	
	// Recreate collections
	tdb.UserCol = tdb.Database.Collection("users")
	tdb.ConvCol = tdb.Database.Collection("conversations")
	tdb.MsgCol = tdb.Database.Collection("messages")
	
	return nil
}

// CreateTestUser creates a test user and returns it
func (tdb *TestDB) CreateTestUser(email, username, fullname string) (*models.User, error) {
	hashedPassword, err := utils.HashPassword("testpassword")
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:            primitive.NewObjectID(),
		Fullname:      fullname,
		Username:      username,
		Email:         email,
		Password:      &hashedPassword,
		Avatar:        "test-avatar.png",
		IsActive:      true,
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = tdb.UserCol.InsertOne(context.Background(), user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateTestMessage creates a test message between two users
func (tdb *TestDB) CreateTestMessage(senderID, receiverID primitive.ObjectID, message string) (*models.Message, error) {
	msg := &models.Message{
		ID:         primitive.NewObjectID(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		Message:    message,
		IsRead:     false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := tdb.MsgCol.InsertOne(context.Background(), msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// CreateTestConversation creates a test conversation between users
func (tdb *TestDB) CreateTestConversation(participants []primitive.ObjectID, messageIDs []primitive.ObjectID) (*models.Conversation, error) {
	conv := &models.Conversation{
		ID:           primitive.NewObjectID(),
		Participants: participants,
		Messages:     messageIDs,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := tdb.ConvCol.InsertOne(context.Background(), conv)
	if err != nil {
		return nil, err
	}

	return conv, nil
}

// SetupTestEnv sets up environment variables for testing
func SetupTestEnv() {
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("SESSION_SECRET", "test-session-secret")
	os.Setenv("GOOGLE_CLIENT_ID", "test-google-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-google-client-secret")
	os.Setenv("GITHUB_ID", "test-github-id")
	os.Setenv("GITHUB_SECRET", "test-github-secret")
	os.Setenv("BASE_URL", "http://localhost:8085")
	os.Setenv("CLIENT_BASE_URL", "http://localhost:3000")
	os.Setenv("ROBOHASH", "https://robohash.org/")
	os.Setenv("PUSHER_APP_ID", "test-pusher-app-id")
	os.Setenv("PUSHER_KEY", "test-pusher-key")
	os.Setenv("PUSHER_SECRET", "test-pusher-secret")
	os.Setenv("PUSHER_CLUSTER", "test-cluster")
}

// GenerateTestJWT generates a JWT token for testing
func GenerateTestJWT(userID, email string) (string, error) {
	return utils.GenerateJWT(userID, email)
}

// AssertSuccessResponse checks if response matches success format
func AssertSuccessResponse(t *testing.T, data map[string]interface{}, expectedMessage string) {
	assert.True(t, data["success"].(bool))
	assert.Equal(t, expectedMessage, data["message"])
	assert.Contains(t, data, "data")
}

// AssertErrorResponse checks if response matches error format  
func AssertErrorResponse(t *testing.T, data map[string]interface{}, expectedMessage string) {
	assert.False(t, data["success"].(bool))
	assert.Equal(t, expectedMessage, data["message"])
	assert.Contains(t, data, "status_code")
}

// RunTestWithDB is a helper that sets up and tears down test database
func RunTestWithDB(t *testing.T, testFunc func(*TestDB)) {
	// Setup test environment
	SetupTestEnv()
	
	// Setup test database
	testDB, err := SetupTestDB()
	if err != nil {
		log.Fatal("Failed to setup test database:", err)
	}
	defer testDB.Cleanup()

	// Run the test
	testFunc(testDB)
}