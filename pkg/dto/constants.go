package dto

import "errors"

// HTTP Header constants
const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// UserIDHeader is the header name for user ID
	UserIDHeader = "X-User-ID"
)

// Context key constants
const (
	// RequestIDKey is the context key for storing request ID
	RequestIDKey = "request_id"
	// UserIDKey is the context key for storing user ID
	UserIDKey = "user_id"
	// UserKey is the context key for storing user object
	UserKey = "user"
)

// Error variables
var (
	// ErrUserNotFound is returned when user is not found in database
	ErrUserNotFound = errors.New("user not found")
	// ErrInvalidUserID is returned when user ID header is invalid
	ErrInvalidUserID = errors.New("invalid user ID")
)
