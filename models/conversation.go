package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Participant struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Avatar    string             `bson:"avatar,omitempty" json:"avatar"`
	Email     string             `bson:"email,omitempty" json:"email"`
	Fullname  string             `bson:"fullname,omitempty" json:"fullname"`
	Username  string             `bson:"username,omitempty" json:"username"`
	CreatedAt time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type Conversation struct {
	ID             primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Participants   []primitive.ObjectID `bson:"participant,omitempty" json:"participant"`
	Messages       []Message            `bson:"messages,omitempty" json:"messages"`
	UnreadMessages int32                `bson:"unreadMessages,omitempty" json:"unread_messages"`
	CreatedAt      time.Time            `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt      time.Time            `bson:"updatedAt,omitempty" json:"updatedAt"`
}
