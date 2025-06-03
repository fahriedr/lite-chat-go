package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID           primitive.ObjectID   `bson:"_id,omitempty" json:"_id"`
	Participants []primitive.ObjectID `bson:"participants,omitempty" json:"participants"`
	Messages     []primitive.ObjectID `bson:"messages,omitempty" json:"messages"`
	CreatedAt    time.Time            `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt    time.Time            `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type ConversationWithSingleParticipant struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Participant UserPublic         `bson:"participants,omitempty" json:"participants"`
	Messages    []Message          `bson:"messages,omitempty" json:"messages"`
	CreatedAt   time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}
