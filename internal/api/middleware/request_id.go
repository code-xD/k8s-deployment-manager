package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for storing request ID
	RequestIDKey = "request_id"
)

// RequestIDMiddleware extracts X-Request-ID from headers and stores it in context
// This middleware is used for idempotency - ensures the same request ID can be used multiple times
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		
		if requestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Request-ID header is required",
			})
			c.Abort()
			return
		}

		// Store request ID in context for later use
		c.Set(RequestIDKey, requestID)
		c.Next()
	}
}
