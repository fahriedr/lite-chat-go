package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SenderID   string             `bson:"senderId,omitempty" json:"senderId"`
	ReceiverID string             `bson:"receiverId,omitempty" json:"receiverid"`
	Message    string             `bson:"message,omitempty" json:"message"`
	IsRead     bool               `bson:"isRead,omitempty" json:"isRead"`
	CreatedAt  time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}
