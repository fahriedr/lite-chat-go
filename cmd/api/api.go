package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type APIServer struct {
	addr string
	db   *mongo.Database
}

func NewAPIServer(addr string, dbURI string, dbName string) (*APIServer, error) {

	clientOptions := options.Client().ApplyURI(dbURI)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	return &APIServer{
		addr: addr,
		db:   db,
	}, nil
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()

	log.Println("Listening on", s.addr)
	return http.ListenAndServe(s.addr, router)
}
