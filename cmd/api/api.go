package api

import (
	"fmt"
	"lite-chat-go/service/conversation"
	"lite-chat-go/service/user"
	"lite-chat-go/utils"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIServer struct {
	userCollection         *mongo.Collection
	conversationCollection *mongo.Collection
	dbName                 string
	port                   string
}

func NewAPIServer(userCollection *mongo.Collection, conversationCollection *mongo.Collection, dbName string, port string) *APIServer {

	return &APIServer{
		userCollection:         userCollection,
		conversationCollection: conversationCollection,
		dbName:                 dbName,
		port:                   port,
	}
}

func (s *APIServer) Run() error {
	mainRouter := mux.NewRouter()
	router := mainRouter.PathPrefix("/api").Subrouter()
	router.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)

	//User route
	userService := user.NewUserService(s.userCollection)
	userRouter := router.PathPrefix("/user").Subrouter()
	userService.RegisterRoutes(userRouter)

	conversationService := conversation.NewConversationService(s.conversationCollection)
	conversationRouter := router.PathPrefix("/conversations").Subrouter()
	conversationService.RegisterRoutes(conversationRouter)

	log.Println("Listening on", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), router)
}

func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]any{"message": "Status OK"})
}
