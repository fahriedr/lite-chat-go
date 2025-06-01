package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Participant struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Fullname      string             `bson:"fullname,omitempty" json:"fullname"`
	Username      string             `bson:"username,omitempty" json:"username"`
	Email         string             `bson:"email,omitempty" json:"email"`
	Avatar        string             `bson:"avatar,omitempty" json:"avatar"`
	EmailVerified bool               `bson:"email_verified,omitempty" json:"email_verified"`
	Password      *string            `bson:"password,omitempty" json:"-"`
	GoogleId      *string            `bson:"google_id,omitempty" json:"google_id"`
	GithubId      *string            `bson:"github_id,omitempty" json:"github_id"`
	AccessToken   *string            `bson:"access_token,omitempty" json:"access_token"`
	Provider      *AuthProvider      `bson:"provider,omitempty" json:"provider"`
	IsActive      bool               `bson:"is_active,omitempty" json:"is_active"`
	CreatedAt     time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}
