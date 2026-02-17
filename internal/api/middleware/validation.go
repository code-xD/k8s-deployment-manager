package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidateRequest validates the request body against the DTO schema
// Returns 422 (Unprocessable Entity) if validation fails
// Works with member functions by accepting a handler function
func ValidateRequest[T any](handler func(c *gin.Context, req *T)) gin.HandlerFunc {
	validate := validator.New()

	return func(c *gin.Context) {
		var req T

		// Bind JSON body to the request struct
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":   "Invalid request body",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate the struct using validator tags
		if err := validate.Struct(&req); err != nil {
			errors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				field := err.Field()
				tag := err.Tag()
				errors[field] = getValidationErrorMessage(field, tag)
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":   "Validation failed",
				"details": errors,
			})
			c.Abort()
			return
		}

		// Call the handler with validated request
		handler(c, &req)
	}
}

// NoBodyHandler wraps a handler that doesn't require a request body
func NoBodyHandler(handler func(c *gin.Context)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c)
	}
}

// getValidationErrorMessage returns a user-friendly error message for validation errors
func getValidationErrorMessage(field, tag string) string {
	switch tag {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " is too short"
	case "max":
		return field + " is too long"
	case "gte":
		return field + " must be greater than or equal to the specified value"
	case "lte":
		return field + " must be less than or equal to the specified value"
	default:
		return field + " failed validation: " + tag
	}
}
