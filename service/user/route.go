package user

import (
	"encoding/json"
	"fmt"
	"lite-chat-go/config"
	"lite-chat-go/models"
	"lite-chat-go/types"
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
	router.HandleFunc("/profile", utils.WithJwtAuth(s.profile)).Methods(http.MethodGet)
}

func (s *UserService) profile(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	email := r.Context().Value(types.ContextKeyEmail).(string)

	var user models.User

	filter := bson.M{
		"email": email,
	}

	err := s.userCollection.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"message": "Login", "data": user})
}

func (s *UserService) handleLogin(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var payload models.UserLoginPayload

	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		log.Println(err)
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User

	filter := bson.M{
		"email": payload.Email,
	}

	err = s.userCollection.FindOne(ctx, filter).Decode(&user)

	if err != nil {
		fmt.Println(err)
		utils.WriteError(w, http.StatusNotFound, "Email or Password are incorrect")
		return
	}

	if user.Password == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Sign in with %s to continue", *user.Provider))
		return
	}

	if !utils.CheckPasswordHash(payload.Password, *user.Password) {
		fmt.Println(err)
		utils.WriteError(w, http.StatusNotFound, "Email or Password are incorrect")
		return
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"success": "Login",
		"user":    user,
		"token":   token,
	})
}

func (s *UserService) handleRegister(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var payload models.UserRegisterPayload
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

	var existingUser models.User

	filter := bson.M{
		"$or": []bson.M{
			{"email": payload.Email},
			{"username": payload.Username},
		},
	}
	err = s.userCollection.FindOne(ctx, filter).Decode(&existingUser)

	if err == nil {
		if existingUser.Email == payload.Email {
			utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("user with email %s already exists", payload.Email))
			return
		}
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("username %s is already taken", payload.Username))
		return
	} else if err != mongo.ErrNoDocuments {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	hashedPassword, err := utils.HashPassword(payload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
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
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	projection := bson.M{
		"fullname": 1,
		"username": 1,
		"email":    1,
		"avatar":   1,
	}

	opts := options.FindOne().SetProjection(projection)

	var insertedDoc models.UserPublic

	filter = bson.M{"_id": res.InsertedID}

	err = s.userCollection.FindOne(ctx, filter, opts).Decode(&insertedDoc)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(insertedDoc.ID.Hex(), insertedDoc.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Login successful",
		"token":   token,
		"user":    insertedDoc,
	})
}
