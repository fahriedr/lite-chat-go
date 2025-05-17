package user

import (
	"encoding/json"
	"fmt"
	"lite-chat-go/config"
	"lite-chat-go/models"
	"lite-chat-go/utils"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserService struct {
	userCollection *mongo.Collection
}

func NewUserService(userCollection *mongo.Collection) *UserService {
	return &UserService{
		userCollection: userCollection,
	}
}

func (s *UserService) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/login", s.handleLogin).Methods(http.MethodPost)
	router.HandleFunc("/register", s.handleRegister).Methods(http.MethodPost)
}

func (s *UserService) handleLogin(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, map[string]any{"message": "Login"})
}

func (s *UserService) handleRegister(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var payload models.UserRegisterPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		log.Fatal(err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	err = utils.Validate.Struct(payload)

	if err != nil {
		log.Fatal(err)
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	var existingUser models.User
	filter := bson.M{"email": payload.Email, "username": payload.Username}
	err = s.userCollection.FindOne(ctx, filter).Decode(&existingUser)

	if err == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("user with email %s already exists", payload.Email))
		return
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	doc := models.User{
		Fullname:      payload.Fullname,
		Username:      payload.Username,
		Email:         payload.Email,
		Password:      &hashedPassword,
		Provider:      nil,
		Avatar:        config.Envs.Robohash + payload.Username,
		IsActive:      true,
		EmailVerified: false,
		GoogleId:      nil,
		GithubId:      nil,
		AccessToken:   nil,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	res, err := s.userCollection.InsertOne(ctx, doc)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	projection := bson.M{
		"fullname": 1,
		"username": 1,
		"email":    1,
		"avatar":   1,
	}

	opts := options.FindOne().SetProjection(projection)

	var insertedDoc models.User

	filter = bson.M{"_id": res.InsertedID}

	err = s.userCollection.FindOne(ctx, filter, opts).Decode(&insertedDoc)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(insertedDoc.ID.Hex(), insertedDoc.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	insertedDoc.Password = nil

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Login successful",
		"token":   token,
		"user":    insertedDoc,
	})
}
