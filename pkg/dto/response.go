package dto

// Standard response structures

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
