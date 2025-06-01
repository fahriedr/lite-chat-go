package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Conversation struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Participants   []primitive.ObjectID `bson:"participants,omitempty" json:"participants"`
	Messages       []Message            `bson:"messages,omitempty" json:"messages"`
	UnreadMessages int32                `bson:"unreadMessages,omitempty" json:"unread_messages"`
	CreatedAt      time.Time            `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type ConversationWithSingleParticipant struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Participant    Participant        `bson:"participants,omitempty" json:"participants"`
	Messages       []Message          `bson:"messages,omitempty" json:"messages"`
	UnreadMessages int32              `bson:"unreadMessages,omitempty" json:"unread_messages"`
	CreatedAt      time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}
