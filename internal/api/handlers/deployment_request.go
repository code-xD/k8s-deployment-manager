package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/code-xd/k8s-deployment-manager/internal/api/middleware"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	portsapi "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/apiService"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DeploymentRequestHandler handles deployment request-related requests
type DeploymentRequestHandler struct {
	deploymentRequest     portsapi.DeploymentRequest
	userRepo              portsdb.User
	deploymentRequestRepo portsdb.DeploymentRequest
	log                   *zap.Logger
}

// NewDeploymentRequestHandler creates a new DeploymentRequestHandler instance with injected dependencies
func NewDeploymentRequestHandler(
	deploymentRequest portsapi.DeploymentRequest,
	userRepo portsdb.User,
	deploymentRequestRepo portsdb.DeploymentRequest,
	log *zap.Logger,
) *DeploymentRequestHandler {
	return &DeploymentRequestHandler{
		deploymentRequest:     deploymentRequest,
		userRepo:              userRepo,
		deploymentRequestRepo: deploymentRequestRepo,
		log:                   log,
	}
}

// GetRoutes returns all deployment request route definitions
func (h *DeploymentRequestHandler) GetRoutes() []dto.RouteDefinition {
	return []dto.RouteDefinition{
		{
			Method: "GET",
			Path:   dto.PathDeploymentRequestsList,
			Middlewares: []gin.HandlerFunc{
				middleware.AuthReadMiddleware(
					h.userRepo,
					h.log,
				),
			},
			Handler: middleware.NoBodyHandler(h.ListDeploymentRequests),
		},
		{
			Method: "GET",
			Path:   dto.PathDeploymentRequestByID,
			Middlewares: []gin.HandlerFunc{
				middleware.AuthReadMiddleware(
					h.userRepo,
					h.log,
				),
			},
			Handler: middleware.NoBodyHandler(h.GetDeploymentRequest),
		},
		{
			Method: "POST",
			Path:   dto.PathDeploymentsCreate,
			// Middlewares are applied in order: RequestID -> Auth -> Validation -> Handler
			// ValidateRequest wraps the handler and provides validated request body
			Middlewares: []gin.HandlerFunc{
				middleware.RequestIDMiddleware(
					h.deploymentRequestRepo,
				),
				middleware.AuthReadWriteMiddleware(
					h.userRepo,
					h.log,
				),
			},
			// Handler wrapped with ValidateRequest to get validated body directly
			Handler: middleware.ValidateRequest[dto.CreateDeploymentRequestWithMetadata](
				h.CreateDeploymentRequest,
			),
		},
		{
			Method: "PATCH",
			Path:   dto.PathDeploymentRequestByID,
			// Middlewares are applied in order: RequestID -> Auth -> Validation -> Handler
			Middlewares: []gin.HandlerFunc{
				middleware.RequestIDMiddleware(
					h.deploymentRequestRepo,
				),
				middleware.AuthReadMiddleware(
					h.userRepo,
					h.log,
				),
			},
			// Handler wrapped with ValidateRequest to get validated body directly
			Handler: middleware.ValidateRequest[dto.UpdateDeploymentRequestMetadata](
				h.UpdateDeploymentRequest,
			),
		},
		{
			Method: "DELETE",
			Path:   dto.PathDeploymentRequestByID,
			// Middlewares are applied in order: RequestID -> Auth -> Handler
			Middlewares: []gin.HandlerFunc{
				middleware.RequestIDMiddleware(
					h.deploymentRequestRepo,
				),
				middleware.AuthReadMiddleware(
					h.userRepo,
					h.log,
				),
			},
			Handler: middleware.NoBodyHandler(h.DeleteDeploymentRequest),
		},
	}
}

// ListDeploymentRequests handles GET /api/v1/deployments/requests
// @Summary      List deployment requests for the authenticated user
// @Description  Returns all deployment requests for the user identified by X-User-ID header
// @Tags         DeploymentRequestService
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string  true  "User ID for authentication"
// @Success      200        {object}  dto.SuccessResponse{data=[]dto.DeploymentRequestListResponse}
// @Failure      401        {object}  dto.ErrorResponse  "Missing or invalid X-User-ID"
// @Failure      403        {object}  dto.ErrorResponse  "User not found"
// @Router       /deployments/requests [get]
func (h *DeploymentRequestHandler) ListDeploymentRequests(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   dto.ErrMsgUserIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	requests, err := h.deploymentRequest.ListDeploymentRequests(c.Request.Context(), userID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   dto.ErrMsgFailedToListDeploymentRequests,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: dto.MsgDeploymentRequestsRetrieved,
		Data:    requests,
	})
}

// GetDeploymentRequest handles GET /api/v1/deployment/requests/:id
// @Summary      Get a deployment request by request ID
// @Description  Returns the full deployment request including metadata for the given request_id. Only returns if owned by the authenticated user.
// @Tags         DeploymentRequestService
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string  true   "User ID for authentication"
// @Param        id         path      string  true   "Request ID of the deployment request"
// @Success      200        {object}  dto.SuccessResponse{data=dto.DeploymentRequestResponse}
// @Failure      401        {object}  dto.ErrorResponse  "Missing or invalid X-User-ID"
// @Failure      403        {object}  dto.ErrorResponse  "User not found"
// @Failure      404        {object}  dto.ErrorResponse  "Deployment request not found"
// @Router       /deployments/requests/{id} [get]
func (h *DeploymentRequestHandler) GetDeploymentRequest(c *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   dto.ErrMsgUserIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	requestID := c.Param(dto.ParamID)
	if requestID == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   dto.ErrMsgRequestIDRequired,
			Details: map[string]interface{}{dto.ResponseKeyParam: dto.ParamID},
		})
		return
	}

	dr, err := h.deploymentRequest.GetDeploymentRequest(c.Request.Context(), requestID, userID.String())
	if err != nil {
		if errors.Is(err, dto.ErrDeploymentRequestNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   dto.ErrMsgDeploymentRequestNotFound,
				Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   dto.ErrMsgFailedToGetDeploymentRequest,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: dto.MsgDeploymentRequestRetrieved,
		Data:    dr,
	})
}

// CreateDeploymentRequest handles POST /api/v1/deployments/requests/create
// @Summary      Create a deployment request
// @Description  Create a new deployment request with metadata
// @Tags         DeploymentRequestService
// @Accept       json
// @Produce      json
// @Param        X-Request-ID  header    string                              true  "Request ID for idempotency"
// @Param        X-User-ID     header    string                              true  "User ID for authentication"
// @Param        request       body      dto.CreateDeploymentRequestWithMetadata  true  "Deployment request details"
// @Success      201           {object}  dto.SuccessResponse{data=dto.DeploymentRequestResponse}
// @Router       /deployments/requests/create [post]
// Request body is validated and provided by ValidateRequest middleware
// RequestID and UserID are available in context from previous middlewares
func (h *DeploymentRequestHandler) CreateDeploymentRequest(c *gin.Context, req *dto.CreateDeploymentRequestWithMetadata) {
	// Extract request ID from context (set by RequestIDMiddleware)
	requestID, err := middleware.GetRequestIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   dto.ErrMsgRequestIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	// Extract user ID from context (set by AuthReadWriteMiddleware)
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   dto.ErrMsgUserIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	// Call service to create deployment request
	deploymentRequest, err := h.deploymentRequest.CreateDeploymentRequest(
		c.Request.Context(),
		req,
		requestID,
		userID.String(),
	)
	if err != nil {
		// Check if it's a conflict error (deployment already exists)
		if err.Error() != "" && strings.Contains(err.Error(), dto.StrAlreadyExists) {
			c.JSON(http.StatusConflict, dto.ErrorResponse{
				Error:   dto.ErrMsgDeploymentAlreadyExists,
				Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   dto.ErrMsgFailedToCreateDeploymentRequest,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{
		Message: dto.MsgDeploymentRequestCreated,
		Data:    deploymentRequest,
	})
}

// UpdateDeploymentRequest handles PATCH /api/v1/deployments/requests/:id
// @Summary      Update a deployment request
// @Description  Update an existing deployment request with optional metadata fields
// @Tags         DeploymentRequestService
// @Accept       json
// @Produce      json
// @Param        X-Request-ID  header    string                              true  "Request ID for idempotency"
// @Param        X-User-ID     header    string                              true  "User ID for authentication"
// @Param        id            path      string                              true  "Deployment identifier"
// @Param        request       body      dto.UpdateDeploymentRequestMetadata true  "Deployment request update details"
// @Success      200           {object}  dto.SuccessResponse{data=dto.DeploymentRequestResponse}
// @Failure      400           {object}  dto.ErrorResponse  "Invalid request"
// @Failure      401           {object}  dto.ErrorResponse  "Missing or invalid X-User-ID"
// @Failure      404           {object}  dto.ErrorResponse  "Deployment not found"
// @Router       /deployments/requests/{id} [patch]
// Request body is validated and provided by ValidateRequest middleware
// RequestID and UserID are available in context from previous middlewares
func (h *DeploymentRequestHandler) UpdateDeploymentRequest(c *gin.Context, req *dto.UpdateDeploymentRequestMetadata) {
	// Extract request ID from context (set by RequestIDMiddleware)
	requestID, err := middleware.GetRequestIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   dto.ErrMsgRequestIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	// Extract user ID from context (set by AuthReadMiddleware)
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   dto.ErrMsgUserIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	// Extract identifier from path parameter
	identifier := c.Param(dto.ParamID)
	if identifier == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   dto.ErrMsgIdentifierRequired,
			Details: map[string]interface{}{dto.ResponseKeyParam: dto.ParamID},
		})
		return
	}

	// Call service to update deployment request
	deploymentRequest, err := h.deploymentRequest.UpdateDeploymentRequest(
		c.Request.Context(),
		identifier,
		req,
		requestID,
		userID.String(),
	)
	if err != nil {
		if errors.Is(err, dto.ErrDeploymentNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   dto.ErrMsgDeploymentNotFound,
				Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to update deployment request",
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: dto.MsgDeploymentRequestUpdated,
		Data:    deploymentRequest,
	})
}

// DeleteDeploymentRequest handles DELETE /api/v1/deployments/requests/:id
// @Summary      Delete a deployment request
// @Description  Delete an existing deployment request
// @Tags         DeploymentRequestService
// @Accept       json
// @Produce      json
// @Param        X-Request-ID  header    string  true  "Request ID for idempotency"
// @Param        X-User-ID     header    string  true  "User ID for authentication"
// @Param        id            path      string  true  "Deployment identifier"
// @Success      200           {object}  dto.SuccessResponse{data=dto.DeploymentRequestResponse}
// @Failure      400           {object}  dto.ErrorResponse  "Invalid request"
// @Failure      401           {object}  dto.ErrorResponse  "Missing or invalid X-User-ID"
// @Failure      404           {object}  dto.ErrorResponse  "Deployment not found"
// @Router       /deployments/requests/{id} [delete]
// RequestID and UserID are available in context from previous middlewares
func (h *DeploymentRequestHandler) DeleteDeploymentRequest(c *gin.Context) {
	// Extract request ID from context (set by RequestIDMiddleware)
	requestID, err := middleware.GetRequestIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   dto.ErrMsgRequestIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	// Extract user ID from context (set by AuthReadMiddleware)
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   dto.ErrMsgUserIDNotFound,
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	// Extract identifier from path parameter
	identifier := c.Param(dto.ParamID)
	if identifier == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   dto.ErrMsgIdentifierRequired,
			Details: map[string]interface{}{dto.ResponseKeyParam: dto.ParamID},
		})
		return
	}

	// Call service to delete deployment request
	deploymentRequest, err := h.deploymentRequest.DeleteDeploymentRequest(
		c.Request.Context(),
		identifier,
		requestID,
		userID.String(),
	)
	if err != nil {
		if errors.Is(err, dto.ErrDeploymentNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error:   dto.ErrMsgDeploymentNotFound,
				Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to delete deployment request",
			Details: map[string]interface{}{dto.ResponseKeyError: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Message: dto.MsgDeploymentRequestDeleted,
		Data:    deploymentRequest,
	})
}
