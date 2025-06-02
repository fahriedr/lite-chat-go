package message

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

type MessageService struct {
	messageCollection      *mongo.Collection
	conversationCollection *mongo.Collection
}

func NewMessageService(messageCollection *mongo.Collection, conversationCollection *mongo.Collection) *MessageService {
	return &MessageService{
		messageCollection:      messageCollection,
		conversationCollection: conversationCollection,
	}
}

func (s *MessageService) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/list/{receiver_id}", utils.WithJwtAuth(s.getMessage)).Methods(http.MethodGet)
}

func (s *MessageService) getMessage(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	receiver_id := mux.Vars(r)["receiver_id"]
	user_id := ctx.Value(types.ContextKeyUserID).(string)

	receiver_id_object, err := primitive.ObjectIDFromHex(receiver_id)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "User ID not found")
		return
	}

	user_id_object, err := primitive.ObjectIDFromHex(user_id)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, " User ID not found")
		return
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "participants", Value: bson.D{
				{Key: "$all", Value: bson.A{user_id_object, receiver_id_object}},
			}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "messages"},
			{Key: "localField", Value: "messages"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "messages"},
		}}},
		bson.D{{Key: "$unwind", Value: "$messages"}},
		bson.D{{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: "$messages"},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "createdAt", Value: -1},
		}}},
		bson.D{{Key: "$limit", Value: 50}},
	}

	cursor, err := s.conversationCollection.Aggregate(ctx, pipeline)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	var message []models.Message

	if err := cursor.All(ctx, &message); err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Cursor error"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.CustomSuccessResponse{
		Success: true,
		Message: "Success",
		Data:    message,
	})
}
