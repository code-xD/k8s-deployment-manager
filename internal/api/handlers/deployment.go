package handlers

import (
	"errors"
	"net/http"

	"github.com/code-xd/k8s-deployment-manager/internal/api/middleware"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	portsapi "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/apiService"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeploymentHandler handles deployment-related requests
type DeploymentHandler struct {
	deploymentService portsapi.Deployment
	userRepo          portsdb.User
	log               *zap.Logger
}

// NewDeploymentHandler creates a new DeploymentHandler instance with injected dependencies
func NewDeploymentHandler(
	deploymentService portsapi.Deployment,
	userRepo portsdb.User,
	log *zap.Logger,
) *DeploymentHandler {
	return &DeploymentHandler{
		deploymentService: deploymentService,
		userRepo:          userRepo,
		log:               log,
	}
}

// GetRoutes returns all deployment route definitions
func (h *DeploymentHandler) GetRoutes() []dto.RouteDefinition {
	return []dto.RouteDefinition{
		{
			Method: "GET",
			Path:   dto.PathDeploymentsList,
			Middlewares: []gin.HandlerFunc{
				middleware.AuthReadMiddleware(
					h.userRepo,
					h.log,
				),
			},
			Handler: middleware.NoBodyHandler(h.ListDeployments),
		},
		{
			Method: "GET",
			Path:   dto.PathDeploymentByID,
			Middlewares: []gin.HandlerFunc{
				middleware.AuthReadMiddleware(
					h.userRepo,
					h.log,
				),
			},
			Handler: middleware.NoBodyHandler(h.GetDeployment),
		},
	}
}

// ListDeployments handles GET /api/v1/deployments
// @Summary      List deployments for the authenticated user
// @Description  Returns all deployments for the user identified by X-User-ID header. Returns limited fields: identifier, createdAt, UpdatedAt, status, name, namespace.
// @Tags         DeploymentService
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string  true  "User ID for authentication"
// @Success      200        {object}  dto.SuccessResponse{data=[]dto.DeploymentListResponse}
// @Failure      401        {object}  dto.ErrorResponse  "Missing or invalid X-User-ID"
// @Failure      403        {object}  dto.ErrorResponse  "User not found"
// @Router       /deployments [get]
func (h *DeploymentHandler) ListDeployments(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   dto.ErrMsgUserIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	deployments, err := h.deploymentService.ListDeployments(c.Request.Context(), userID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   dto.ErrMsgFailedToListDeployments,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: dto.MsgDeploymentsRetrieved,
		Data:    deployments,
	})
}

// GetDeployment handles GET /api/v1/deployments/:id
// @Summary      Get a deployment by identifier
// @Description  Returns the full deployment including metadata for the given identifier. Only returns if owned by the authenticated user.
// @Tags         DeploymentService
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string  true   "User ID for authentication"
// @Param        id         path      string  true   "Identifier of the deployment"
// @Success      200        {object}  dto.SuccessResponse{data=dto.DeploymentResponse}
// @Failure      401        {object}  dto.ErrorResponse  "Missing or invalid X-User-ID"
// @Failure      403        {object}  dto.ErrorResponse  "User not found"
// @Failure      404        {object}  dto.ErrorResponse  "Deployment not found"
// @Router       /deployments/{id} [get]
func (h *DeploymentHandler) GetDeployment(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   dto.ErrMsgUserIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	identifier := c.Param(dto.ParamID)
	if identifier == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   dto.ErrMsgIdentifierRequired,
			Details: map[string]interface{}{dto.ResponseKeyParam: dto.ParamID},
		})
		return
	}

	deployment, err := h.deploymentService.GetDeployment(c.Request.Context(), identifier, userID.String())
	if err != nil {
		if errors.Is(err, dto.ErrDeploymentNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   dto.ErrMsgDeploymentNotFound,
				Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   dto.ErrMsgFailedToGetDeployment,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: dto.MsgDeploymentRetrieved,
		Data:    deployment,
	})
}
