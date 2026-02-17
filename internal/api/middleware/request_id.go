package middleware

import (
	"net/http"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsrepo "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo"
	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware extracts X-Request-ID from headers and stores it in context
// This middleware is used for idempotency - ensures the same request ID can be used multiple times
func RequestIDMiddleware(
	deploymentRequestRepo portsrepo.DeploymentRequest,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(dto.RequestIDHeader)

		if requestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "X-Request-ID header is required",
			})
			c.Abort()
			return
		}
		_, found, err := deploymentRequestRepo.GetByRequestID(c.Request.Context(), requestID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check existing deployment request",
			})
			c.Abort()
			return
		}

		if found {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Deployment request with same request ID already exists",
			})
			c.Abort()
			return
		}

		// Store request ID in context for later use
		c.Set(dto.RequestIDKey, requestID)
		c.Next()
	}
}
