package main

import (
	"context"
	"fmt"
	"lite-chat-go/cmd/api"
	"lite-chat-go/config"
	"lite-chat-go/utils"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	mongoClient            *mongo.Client
	userCollection         *mongo.Collection
	conversationCollection *mongo.Collection
	messageCollection      *mongo.Collection
)

func init() {
	ctx := context.TODO()

	dbUri := config.Envs.MongoUrl
	dbName := config.Envs.Database
	connectionOpts := options.Client().ApplyURI(dbUri)

	mongoClient, err := mongo.Connect(ctx, connectionOpts)

	if err != nil {
		fmt.Printf("an error ocurred when connect to mongoDB : %v", err)
		log.Fatal(err)
	}

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	database := mongoClient.Database(dbName)

	log.Println("MongoDB successfully connected")
	userCollection = database.Collection("users")
	conversationCollection = database.Collection("conversations")
	messageCollection = database.Collection("messages")

	// Drop existing googleId index if it exists
	indexes, err := userCollection.Indexes().List(ctx)
	if err != nil {
		log.Fatalf("Error listing indexes: %v", err)
	}

	for indexes.Next(ctx) {
		var index bson.M
		if err := indexes.Decode(&index); err != nil {
			log.Fatalf("Error decoding index: %v", err)
		}
		if key, ok := index["key"].(bson.M); ok {
			if _, exists := key["googleId"]; exists {
				if name, ok := index["name"].(string); ok {
					fmt.Printf("Dropping index: %s\n", name)
					if _, err := userCollection.Indexes().DropOne(ctx, name); err != nil {
						log.Fatalf("Error dropping index: %v", err)
					}
				}
			}
		}
	}

	// Create sparse unique index on googleId
	mod := mongo.IndexModel{
		Keys:    bson.D{{Key: "googleId", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true).SetName("googleId_sparse_unique"),
	}

	_, err = userCollection.Indexes().CreateOne(ctx, mod)
	if err != nil {
		log.Fatalf("Failed to create sparse unique index on googleId: %v", err)
	}

	fmt.Println("✅ Sparse unique index on googleId created successfully")
}

func main() {

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("Starting Lite Chat API server...",
		zap.String("environment", config.Envs.Environment),
		zap.Int("port", 8080),
	)

	utils.InitLogger(logger)

	server := api.NewAPIServer(
		userCollection,
		conversationCollection,
		messageCollection,
		config.Envs.Database,
		config.Envs.Port,
		logger,
	)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
