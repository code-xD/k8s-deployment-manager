package middleware

import (
	"errors"
	"net/http"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsrepo "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AuthReadMiddleware validates that the user exists in the database
// Returns 401 if header is missing/invalid, 403 if user not found
func AuthReadMiddleware(
	userRepo portsrepo.User,
	logger *zap.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userExternalID := c.GetHeader(dto.UserIDHeader)

		if userExternalID == "" {
			logger.Warn("X-User-ID header is missing")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "X-User-ID header is required",
			})
			c.Abort()
			return
		}

		// Get user from database
		user, err := userRepo.GetByExternalID(c.Request.Context(), userExternalID)
		if err != nil {
			logger.Warn("User not found", zap.String("external_id", userExternalID), zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User not found",
			})
			c.Abort()
			return
		}

		// Store user ID and user object in context
		c.Set(dto.UserIDKey, user.ID)
		c.Set(dto.UserKey, user)
		c.Next()
	}
}

// AuthReadWriteMiddleware validates user and creates if not exists
// Returns 401 if header is missing/invalid, creates user if not found
func AuthReadWriteMiddleware(
	userRepo portsrepo.User,
	logger *zap.Logger,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userExternalID := c.GetHeader(dto.UserIDHeader)

		if userExternalID == "" {
			logger.Warn("X-User-ID header is missing")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "X-User-ID header is required",
			})
			c.Abort()
			return
		}

		// Try to get user from database
		user, err := userRepo.GetByExternalID(c.Request.Context(), userExternalID)
		if err != nil {
			// User doesn't exist, create it
			logger.Info("User not found, creating new user", zap.String("external_id", userExternalID))
			user = &models.User{
				UserExternalID: userExternalID,
			}

			if err := userRepo.Create(c.Request.Context(), user); err != nil {
				logger.Error("Failed to create user", zap.String("external_id", userExternalID), zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to create user",
				})
				c.Abort()
				return
			}

			logger.Info("User created successfully", zap.String("external_id", userExternalID), zap.String("user_id", user.ID.String()))
		}

		// Store user ID and user object in context
		c.Set(dto.UserIDKey, user.ID)
		c.Set(dto.UserKey, user)
		c.Next()
	}
}

// GetUserIDFromContext extracts user ID from gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get(dto.UserIDKey)
	if !exists {
		return uuid.Nil, dto.ErrInvalidUserID
	}

	id, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil, dto.ErrInvalidUserID
	}

	return id, nil
}

// GetRequestIDFromContext extracts request ID from gin context
func GetRequestIDFromContext(c *gin.Context) (string, error) {
	requestID, exists := c.Get(dto.RequestIDKey)
	if !exists {
		return "", errors.New("request ID not found in context")
	}

	id, ok := requestID.(string)
	if !ok {
		return "", errors.New("invalid request ID type in context")
	}

	return id, nil
}
