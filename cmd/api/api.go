package api

import (
	"fmt"
	"lite-chat-go/config"
	"lite-chat-go/service/conversation"
	"lite-chat-go/service/message"
	"lite-chat-go/service/user"
	"lite-chat-go/utils"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"go.mongodb.org/mongo-driver/mongo"
)

type APIServer struct {
	userCollection         *mongo.Collection
	conversationCollection *mongo.Collection
	messageCollection      *mongo.Collection
	dbName                 string
	port                   string
}

var store = sessions.NewCookieStore([]byte(config.Envs.SessionSecret))

func NewAPIServer(userCollection *mongo.Collection, conversationCollection *mongo.Collection, messageCollection *mongo.Collection, dbName string, port string) *APIServer {

	return &APIServer{
		userCollection:         userCollection,
		conversationCollection: conversationCollection,
		messageCollection:      messageCollection,
		dbName:                 dbName,
		port:                   port,
	}
}

func (s *APIServer) Run() error {
	gothic.Store = store
	goth.UseProviders(
		google.New(config.Envs.GoogleClientID, config.Envs.GoogleClientSecret, fmt.Sprintf("%s/api/user/auth/google/callback", config.Envs.BaseUrl), "email", "profile"),
		github.New(config.Envs.GithubId, config.Envs.GithubSecret, fmt.Sprintf("%s/api/user/auth/github/callback", config.Envs.BaseUrl), "user:email"),
	)

	mainRouter := mux.NewRouter()
	router := mainRouter.PathPrefix("/api").Subrouter()
	router.HandleFunc("/health", s.healthCheck).Methods(http.MethodGet)

	//User route
	userService := user.NewUserService(s.userCollection)
	userRouter := router.PathPrefix("/user").Subrouter()
	userService.RegisterRoutes(userRouter)

	//Conversation route
	conversationService := conversation.NewConversationService(s.conversationCollection)
	conversationRouter := router.PathPrefix("/conversations").Subrouter()
	conversationService.RegisterRoutes(conversationRouter)

	//Message route
	messageService := message.NewMessageService(s.messageCollection, s.conversationCollection, s.userCollection)
	messageRouter := router.PathPrefix("/messages").Subrouter()
	messageService.RegisterRoutes(messageRouter)

	// CORS config
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:5173"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization", "Access-Control-Allow-Origin"})

	// Apply CORS middleware
	handler := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(mainRouter)

	log.Println("Listening on", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%s", s.port), handler)
}

func (s *APIServer) healthCheck(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]any{"message": "Status OK"})
}
