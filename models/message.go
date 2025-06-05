package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	SenderID   primitive.ObjectID `bson:"senderId,omitempty" json:"senderId"`
	ReceiverID primitive.ObjectID `bson:"receiverId,omitempty" json:"receiverid"`
	Message    string             `bson:"message,omitempty" json:"message"`
	IsRead     bool               `bson:"isRead" json:"isRead"`
	CreatedAt  time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type MessagePayload struct {
	UserId  string `json:"userId" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type UpdateMessagePayload struct {
	MessageID string `json:"messageId" validate:"required"`
}
