package conversation

import (
	"lite-chat-go/models"
	"lite-chat-go/types"
	"lite-chat-go/utils"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConversationService struct {
	conversationCollection *mongo.Collection
}

func NewConversationService(conversationCollection *mongo.Collection) *ConversationService {
	return &ConversationService{
		conversationCollection: conversationCollection,
	}
}

func (s *ConversationService) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("", utils.WithJwtAuth( s.getConversation)).Methods(http.MethodGet)
}

func (s *ConversationService) getConversation(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()

	userId := r.Context().Value(types.ContextKeyUserID).(string)

	var conversations models.Conversation

	filter := bson.M{
		"_id": userId,
	}
	utils.WriteJSON(w, http.StatusOK, map[string]any{"message": "Conversation"})
}
