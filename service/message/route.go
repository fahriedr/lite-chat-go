package message

import (
	"encoding/json"
	"lite-chat-go/config"
	"lite-chat-go/models"
	"lite-chat-go/types"
	"lite-chat-go/utils"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pusher/pusher-http-go/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var client = pusher.Client{
	AppID:   config.Envs.PusherAppID,
	Key:     config.Envs.PusherKey,
	Secret:  config.Envs.PusherSecret,
	Cluster: config.Envs.PusherCluster,
	Secure:  true,
}

type MessageService struct {
	messageCollection      *mongo.Collection
	conversationCollection *mongo.Collection
	userCollection         *mongo.Collection
}

func NewMessageService(messageCollection *mongo.Collection, conversationCollection *mongo.Collection, userCollection *mongo.Collection) *MessageService {
	return &MessageService{
		messageCollection:      messageCollection,
		conversationCollection: conversationCollection,
		userCollection:         userCollection,
	}
}

func (s *MessageService) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/list/{receiver_id}", utils.WithJwtAuth(s.getMessage)).Methods(http.MethodGet)
	router.HandleFunc("/send", utils.WithJwtAuth(s.sendMessage)).Methods(http.MethodPost)
	router.HandleFunc("/update-status", utils.WithJwtAuth(s.updateStatusMessage)).Methods(http.MethodPost)
}

func (s *MessageService) getMessage(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	receiverId := mux.Vars(r)["receiver_id"]
	userId := ctx.Value(types.ContextKeyUserID).(string)

	receiverIdObject, err := primitive.ObjectIDFromHex(receiverId)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, "User ID not found")
		return
	}

	userIdObject, err := primitive.ObjectIDFromHex(userId)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, " User ID not found")
		return
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "participants", Value: bson.D{
				{Key: "$all", Value: bson.A{userIdObject, receiverIdObject}},
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

	message := make([]models.Message, 0)

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

func (s *MessageService) sendMessage(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()
	userId, err := primitive.ObjectIDFromHex(ctx.Value(types.ContextKeyUserID).(string))

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var payload models.MessagePayload
	err = json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = utils.Validate.Struct(payload)

	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	receiverObjectId, err := primitive.ObjectIDFromHex(payload.UserId)

	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Find receiver
	var receiver models.User
	err = s.userCollection.FindOne(ctx, bson.M{"_id": receiverObjectId}).Decode(&receiver)
	if err == mongo.ErrNoDocuments {
		utils.WriteError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	receiverId := receiver.ID.Hex() // or your custom getPlainId equivalent

	if receiverId == userId.Hex() {
		utils.WriteError(w, http.StatusBadRequest, "User Id not valid")
		return
	}

	newMessage := models.Message{
		SenderID:   userId,
		ReceiverID: receiverObjectId,
		Message:    payload.Message,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	var conversation models.Conversation
	filter := bson.M{"participants": bson.M{"$all": bson.A{userId, receiverObjectId}}}
	err = s.conversationCollection.FindOne(ctx, filter).Decode(&conversation)

	if err == mongo.ErrNoDocuments {
		// Create new conversation
		conversation = models.Conversation{
			Participants: []primitive.ObjectID{userId, receiverObjectId},
			UpdatedAt:    time.Now(),
		}

		result, err := s.conversationCollection.InsertOne(ctx, conversation)
		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		conversation.ID = result.InsertedID.(primitive.ObjectID)
	} else if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Insert message
	msgResult, err := s.messageCollection.InsertOne(ctx, newMessage)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	newMessage.ID = msgResult.InsertedID.(primitive.ObjectID)

	// Update conversation with message reference
	update := bson.M{
		"$push": bson.M{"messages": msgResult.InsertedID},
		"$set":  bson.M{"updatedAt": time.Now()},
	}
	_, err = s.conversationCollection.UpdateByID(ctx, conversation.ID, update)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	client.Trigger("lite-chat", "upcoming-message", newMessage)

	utils.WriteJSON(w, http.StatusOK, types.CustomSuccessResponse{
		Message: "Success",
		Status:  200,
		Success: true,
		Data:    newMessage,
	})
}

func (s *MessageService) updateStatusMessage(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	userId := ctx.Value(types.ContextKeyUserID).(string)

	var payload models.UpdateMessagePayload

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = utils.Validate.Struct(payload)

	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	messageIdObject, err := primitive.ObjectIDFromHex(payload.MessageID)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var message models.Message
	err = s.messageCollection.FindOne(ctx, bson.M{"_id": messageIdObject}).Decode(&message)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if userId != message.ReceiverID.Hex() {
		utils.WriteError(w, http.StatusBadRequest, "Error processing data")
		return
	}

	message.IsRead = true

	update := bson.M{
		"$set": bson.M{"isRead": true, "updatedAt": time.Now()},
	}

	data, err := s.messageCollection.UpdateByID(ctx, message.ID, update)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		types.CustomSuccessResponse{
			Message: "Success Update message status",
			Status:  http.StatusOK,
			Success: true,
			Data:    data,
		},
	)
}
