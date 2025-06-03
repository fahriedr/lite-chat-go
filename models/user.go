package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthProvider string

const (
	ProviderGoogle AuthProvider = "google"
	ProviderGithub AuthProvider = "github"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Fullname      string             `bson:"fullname,omitempty" json:"fullname"`
	Username      string             `bson:"username,omitempty" json:"username"`
	Email         string             `bson:"email,omitempty" json:"email"`
	Avatar        string             `bson:"avatar,omitempty" json:"avatar"`
	EmailVerified bool               `bson:"IsEmailVerified,omitempty" json:"IsEmailVerified"`
	Password      *string            `bson:"password,omitempty" json:"-"`
	GoogleId      *string            `bson:"googleId,omitempty" json:"googleId"`
	GithubId      *string            `bson:"githubId,omitempty" json:"githubId"`
	AccessToken   *string            `bson:"accessToken,omitempty" json:"accessToken"`
	Provider      *AuthProvider      `bson:"provider,omitempty" json:"provider"`
	IsActive      bool               `bson:"isActive,omitempty" json:"isActive"`
	CreatedAt     time.Time          `bson:"createdAt,omitempty" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt,omitempty" json:"updatedAt"`
}

type UserRegisterPayload struct {
	Fullname        string `json:"fullname" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Username        string `json:"username" validate:"required"`
	Password        string `json:"password" validate:"required,min=3,max=130"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,min=3,max=130"`
}

type UserLoginPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserPublic struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Fullname string             `json:"fullname"`
	Username string             `json:"username"`
	Email    string             `json:"email"`
	Avatar   string             `json:"avatar"`
}
