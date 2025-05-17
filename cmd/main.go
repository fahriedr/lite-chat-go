package main

import (
	"context"
	"encoding/json"
	"fmt"
	"lite-chat-go/config"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	taskCollection *mongo.Collection
)

func init() {
	ctx := context.TODO()

	dbUri := config.Envs.MongoUrl
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
}

func main() {
	db, err := 
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	http.ListenAndServe(":8085", r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello World")
	json.NewEncoder(w).Encode("Hello World")
}

