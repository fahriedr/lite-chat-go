package conversation

import (
	"lite-chat-go/models"
	"lite-chat-go/types"
	"lite-chat-go/utils"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	router.HandleFunc("", utils.WithJwtAuth(s.getConversation)).Methods(http.MethodGet)
}

func (s *ConversationService) getConversation(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	userId := r.Context().Value(types.ContextKeyUserID).(string)
	userIdObject, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Failed to fetch User id object")
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{
			"participants": bson.M{"$in": []primitive.ObjectID{userIdObject}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "participants",
			"foreignField": "_id",
			"as":           "participants",
		}}},
		bson.D{{Key: "$set", Value: bson.M{
			"participants": bson.M{
				"$filter": bson.M{
					"input": "$participants",
					"as":    "participant",
					"cond": bson.M{
						"$ne": []interface{}{"$$participant._id", userIdObject},
					},
				},
			},
		}}},
		bson.D{{Key: "$set", Value: bson.M{
			"participants": bson.M{"$first": "$participants"},
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         "messages",
			"localField":   "messages",
			"foreignField": "_id",
			"as":           "messages",
		}}},
		bson.D{{Key: "$project", Value: bson.M{
			"messages":     bson.M{"$slice": []interface{}{"$messages", -1}},
			"participants": 1,
		}}},
	}

	cursor, err := s.conversationCollection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to fetch cursor"})
		return
	}

	var conversations []models.ConversationWithSingleParticipant
	if err := cursor.All(ctx, &conversations); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to decode conversations"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"message": "Success", "data": conversations, "success": true})
}
