package utils

import (
	"encoding/json"
	"fmt"
	"lite-chat-go/types"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

func RandomString(length int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "Application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err string) {
	WriteJSON(
		w,
		status, types.CustomErrorResponse{
			Success:    false,
			StatusCode: status,
			Message:    err,
		},
	)
}

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}

func EmailToUsername(email string) string {
	parts := strings.Split(email, "@")
	localPart := parts[0]
	randomDigits := rand.Intn(9000) + 1000

	return fmt.Sprintf("%s%d", localPart, randomDigits)
}

func MapToJSON(data map[string]interface{}) (string, error) {
	// Marshal the map to JSON
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
