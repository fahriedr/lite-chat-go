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
