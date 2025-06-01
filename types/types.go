package types

type ErrorDetails struct {
	Field   *string
	Code    string
	Message string
}

type CustomErrorResponse struct {
	Success    bool          `json:"success"`
	StatusCode int           `json:"status_code"`
	Message    string        `json:"message"`
	Details    *ErrorDetails `json:"details"`
}

type CustomSuccessResponse struct {
	Message string                  `json:"message"`
	Success bool                    `json:"success"`
	Data    interface{}             `json:"data,omitempty"`
	Status  int                     `json:"status,omitempty"`
	Headers *map[string]interface{} `json:"headers,omitempty"`
}

type contextKey string

const (
	ContextKeyUserID contextKey = "userID"
	ContextKeyEmail  contextKey = "email"
)
