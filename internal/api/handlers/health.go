package handlers

import (
	"net/http"

	"github.com/code-xd/k8s-deployment-manager/internal/api/middleware"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
}

// NewHealthHandler creates a new HealthHandler instance
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// GetRoutes returns all health check route definitions
func (h *HealthHandler) GetRoutes() []dto.RouteDefinition {
	return []dto.RouteDefinition{
		{
			Method:  "GET",
			Path:    "/api/v1/ping",
			Handler: middleware.NoBodyHandler(h.Ping),
		},
	}
}

// Ping handles GET /api/v1/ping
func (h *HealthHandler) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
