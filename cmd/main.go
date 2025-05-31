package main

import (
	"context"
	"fmt"
	"lite-chat-go/cmd/api"
	"lite-chat-go/config"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient            *mongo.Client
	userCollection         *mongo.Collection
	conversationCollection *mongo.Collection
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

	log.Println("MongoDB successfully connected")
	userCollection = mongoClient.Database(dbName).Collection("users")
	conversationCollection = mongoClient.Database(dbName).Collection("conversations")
}

func main() {
	server := api.NewAPIServer(userCollection, conversationCollection, config.Envs.Database, config.Envs.Port)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
