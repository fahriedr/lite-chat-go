package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"lite-chat-go/config"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"lite-chat-go/types"
)

// JWT secret key - should be stored in environment variables in production
var jwtSecret = []byte(config.Envs.JWTSecret)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type Claims struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func GenerateJWT(id, email string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)

	claims := &Claims{
		ID:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func WithJwtAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := getTokenFromRequest(r)
		token, err := ValidateJWT(tokenString)

		if err != nil {
			log.Printf("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		fmt.Println(token.ID, "token")
		ctx := context.WithValue(r.Context(), types.ContextKeyUserID, token.ID)
		ctx = context.WithValue(ctx, types.ContextKeyEmail, token.Email)

		handlerFunc(w, r.WithContext(ctx))
	}
}

func getTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")

	if tokenAuth != "" {
		return tokenAuth
	}

	return ""
}

func permissionDenied(w http.ResponseWriter) {
	writeError(w, http.StatusForbidden, fmt.Errorf("permission denied"))
}

func writeError(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Add("Content-Type", "Application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err)
}
