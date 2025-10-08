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
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	router.HandleFunc("/search/{query}", utils.WithJwtAuth(s.handleSearch)).Methods(http.MethodGet)
	router.HandleFunc("/auth/{provider}", gothic.BeginAuthHandler)
	router.HandleFunc("/auth/{provider}/callback", s.handleAuthProviderCallback).Methods(http.MethodGet, http.MethodPost)
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
		utils.WriteError(w, http.StatusNotFound, "Email or Password are incorrect")
		return
	}

	if user.Password == nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Sprintf("Sign in with %s to continue", *user.Provider))
		return
	}

	if !utils.CheckPasswordHash(payload.Password, *user.Password) {
		utils.WriteError(w, http.StatusNotFound, "Email or Password are incorrect")
		return
	}

	userPublic := models.UserPublic{
		ID:       user.ID,
		Fullname: user.Fullname,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}

	token, err := utils.GenerateJWT(user.ID.Hex(), user.Email)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"success": "Login",
		"user":    userPublic,
		"token":   token,
	})
}

func (s *UserService) handleRegister(w http.ResponseWriter, r *http.Request) {

	var ctx = r.Context()

	var payload models.UserRegisterPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	err = utils.Validate.Struct(payload)

	if err != nil {
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

func (s *UserService) handleSearch(w http.ResponseWriter, r *http.Request) {
	var ctx = r.Context()

	querySearch := mux.Vars(r)["query"]
	userId := ctx.Value(types.ContextKeyUserID).(string)
	userIdObject, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, " User ID not found")
		return
	}

	var user []models.UserPublic

	filter := bson.M{
		"_id": bson.M{"$ne": userIdObject},
		"$or": []bson.M{
			{"username": bson.M{"$regex": querySearch, "$options": "i"}},
			{"email": bson.M{"$regex": querySearch, "$options": "i"}},
		},
	}

	cursor, err := s.userCollection.Find(ctx, filter)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err = cursor.All(ctx, &user); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if user == nil {
		user = []models.UserPublic{}
	}

	utils.WriteJSON(
		w,
		http.StatusOK,
		types.CustomSuccessResponse{
			Message: "Success",
			Status:  http.StatusOK,
			Success: true,
			Data:    user,
		},
	)
}

func (s *UserService) handleAuthProviderCallback(w http.ResponseWriter, r *http.Request) {
	userGoth, err := gothic.CompleteUserAuth(w, r)
	ctx := r.Context()

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var existingUser models.User

	filter := bson.M{
		"$or": []bson.M{
			{"email": userGoth.Email},
		},
	}

	err = s.userCollection.FindOne(ctx, filter).Decode(&existingUser)

	provider := userGoth.Provider
	providerId := userGoth.UserID

	if err != nil {

		username := utils.EmailToUsername(userGoth.Email)

		doc := models.User{
			Fullname:      userGoth.Name,
			Email:         userGoth.Email,
			Username:      username,
			Avatar:        config.Envs.Robohash + username,
			Password:      nil,
			Provider:      (*models.AuthProvider)(&userGoth.Provider),
			EmailVerified: true,
			IsActive:      true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		switch provider {
		case "google":
			doc.GoogleId = &providerId
		case "github":
			doc.GithubId = &providerId
		default:
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, err := s.userCollection.InsertOne(ctx, doc)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err.Error())
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

		userJSON, err := json.Marshal(insertedDoc)
		if err != nil {
			http.Error(w, "failed to encode user", http.StatusInternalServerError)
			return
		}

		encodedUser := url.QueryEscape(string(userJSON))

		http.Redirect(w, r, fmt.Sprintf("%s/oauth/callback?token=%s&user=%s", config.Envs.ClientBaseUrl, token, encodedUser), http.StatusFound)

	} else {

		updated := false

		if existingUser.Provider != (*models.AuthProvider)(&userGoth.Provider) {
			if provider == "google" {
				existingUser.GoogleId = &providerId
			} else {
				existingUser.GithubId = &providerId
			}
			updated = true
		}

		if updated {
			update := bson.M{
				"$set": bson.M{"EmailVerified": true, "updatedAt": time.Now()},
			}

			switch provider {
			case "google":
				userGoth.Provider = "GoogleId"
			case "github":
				userGoth.Provider = "GithubId"
			}

			update["$set"].(bson.M)[userGoth.Provider] = providerId

			_, err := s.userCollection.UpdateByID(ctx, existingUser.ID, update)

			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("unsupported provider: %s", provider))
				return
			}

			// Generate JWT token
			token, err := utils.GenerateJWT(existingUser.ID.Hex(), existingUser.Email)
			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err.Error())
				return
			}

			var insertedDoc models.UserPublic

			filter = bson.M{"_id": existingUser.ID}
			projection := bson.M{
				"fullname": 1,
				"username": 1,
				"email":    1,
				"avatar":   1,
			}

			opts := options.FindOne().SetProjection(projection)

			err = s.userCollection.FindOne(ctx, filter, opts).Decode(&insertedDoc)

			if err != nil {
				utils.WriteError(w, http.StatusInternalServerError, err.Error())
				return
			}

			userJSON, err := json.Marshal(insertedDoc)
			if err != nil {
				http.Error(w, "failed to encode user", http.StatusInternalServerError)
				return
			}

			encodedUser := url.QueryEscape(string(userJSON))

			http.Redirect(w, r, fmt.Sprintf("%s/oauth/callback?token=%s&user=%s", config.Envs.ClientBaseUrl, token, encodedUser), http.StatusFound)
		}

	}
}
