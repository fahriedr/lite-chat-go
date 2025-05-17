package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	_id            primitive.ObjectID `bson:"_id,omitempty"`
	fullname       string             `bson:"fullname,omitempty"`
	username       string             `bson:"username,omitempty"`
	email          string             `bson:"email,omitempty"`
	avatar         string             `bson:"avatar,omitempty`
	email_verified bool               `bson:"email_verified,omitempty`
	password       string             `bson:"password,omitempty`
	google_id      string             `bson:"google_id,omitempty`
	github_id      string             `bson:"github_id,omitempty`
	access_token   string             `bson:"access_token,omitempty`
	provider       string             `bson:"provider,omitempty`
	is_active      string             `bson:"is_active,omitempty`
}

type RegisterUserPayload struct {
	Fullname        string `bson:"fullname" validate:"required"`
	Email           string `bson:"email" validate:"required,email"`
	username        string `bson:"username" validate:"required"`
	Password        string `bson:"password" validate:"required,min=3,max=130"`
	ConfirmPassword string `bson:"Confirm_password" validate:"required,min=3,max=130"`
}
