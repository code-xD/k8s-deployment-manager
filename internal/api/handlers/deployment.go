package handlers

import (
	"net/http"

	"github.com/code-xd/k8s-deployment-manager/internal/api/middleware"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/gin-gonic/gin"
)

// DeploymentHandler handles deployment-related requests
type DeploymentHandler struct {
	// Add any dependencies here (e.g., services, repositories)
}

// NewDeploymentHandler creates a new DeploymentHandler instance
func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{}
}

// GetRoutes returns all deployment route definitions
func (h *DeploymentHandler) GetRoutes() []dto.RouteDefinition {
	return []dto.RouteDefinition{
		{
			Method:  "POST",
			Path:    "/api/v1/deployments",
			Handler: middleware.ValidateRequest[dto.CreateDeploymentRequest](h.CreateDeployment),
		},
		{
			Method:  "PUT",
			Path:    "/api/v1/deployments/:id",
			Handler: middleware.ValidateRequest[dto.UpdateDeploymentRequest](h.UpdateDeployment),
		},
		{
			Method:  "GET",
			Path:    "/api/v1/deployments",
			Handler: middleware.NoBodyHandler(h.ListDeployments),
		},
		{
			Method:  "GET",
			Path:    "/api/v1/deployments/:id",
			Handler: middleware.NoBodyHandler(h.GetDeployment),
		},
	}
}

// CreateDeployment handles POST /api/v1/deployments
// This handler requires a request body and uses validation middleware
func (h *DeploymentHandler) CreateDeployment(c *gin.Context, req *dto.CreateDeploymentRequest) {
	// Handler logic here - req is already validated
	c.JSON(http.StatusCreated, gin.H{
		"message": "Deployment created successfully",
		"data": gin.H{
			"name":      req.Name,
			"namespace": req.Namespace,
			"image":     req.Image,
			"replicas":  req.Replicas,
		},
	})
}

// UpdateDeployment handles PUT /api/v1/deployments/:id
// This handler requires a request body and uses validation middleware
func (h *DeploymentHandler) UpdateDeployment(c *gin.Context, req *dto.UpdateDeploymentRequest) {
	deploymentID := c.Param("id")

	// Handler logic here - req is already validated
	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment updated successfully",
		"data": gin.H{
			"id":       deploymentID,
			"replicas": req.Replicas,
		},
	})
}

// ListDeployments handles GET /api/v1/deployments
// This handler does NOT require a request body
func (h *DeploymentHandler) ListDeployments(c *gin.Context) {
	// Handler logic here - no body validation needed
	c.JSON(http.StatusOK, gin.H{
		"message": "Deployments retrieved successfully",
		"data":    []string{"deployment-1", "deployment-2"},
	})
}

// GetDeployment handles GET /api/v1/deployments/:id
// This handler does NOT require a request body
func (h *DeploymentHandler) GetDeployment(c *gin.Context) {
	deploymentID := c.Param("id")

	// Handler logic here - no body validation needed
	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment retrieved successfully",
		"data": gin.H{
			"id":   deploymentID,
			"name": "example-deployment",
		},
	})
}
