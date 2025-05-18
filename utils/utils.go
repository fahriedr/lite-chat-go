package utils

import (
	"encoding/json"
	"fmt"
	"lite-chat-go/types"
	"math/rand"
	"net/http"

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
	fmt.Println(v, "v")
	w.Header().Add("Content-Type", "Application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, err string) {
	WriteJSON(
		w,
		status, map[string]types.CustomErrorResponse{
			"error": {
				Success:    false,
				StatusCode: status,
				Message:    err,
			},
		},
	)
}

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)
}
