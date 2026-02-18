package apiService

import (
	"context"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	portsqueue "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/queue"
	portsapi "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/apiService"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeploymentRequestService implements the deployment request business logic for the API
type DeploymentRequestService struct {
	repo           portsdb.DeploymentRequest
	deploymentRepo portsdb.Deployment
	publisher      portsqueue.DeploymentRequest
	logger         *zap.Logger
}

// NewDeploymentRequestService creates a new DeploymentRequestService with injected dependencies
func NewDeploymentRequestService(
	repo portsdb.DeploymentRequest,
	deploymentRepo portsdb.Deployment,
	publisher portsqueue.DeploymentRequest,
	logger *zap.Logger,
) portsapi.DeploymentRequest {
	return &DeploymentRequestService{
		repo:           repo,
		deploymentRepo: deploymentRepo,
		publisher:      publisher,
		logger:         logger,
	}
}

// CreateDeploymentRequest handles the business logic for creating a deployment request
func (s *DeploymentRequestService) CreateDeploymentRequest(
	ctx context.Context,
	req *dto.CreateDeploymentRequestWithMetadata,
	requestID string,
	userID string,
) (*dto.DeploymentRequestResponse, error) {
	s.logger.Info("Creating deployment request",
		zap.String("request_id", requestID),
		zap.String("name", req.Name),
		zap.String("namespace", req.Namespace),
		zap.String("user_id", userID),
	)

	// Step 1: Check if deployment exists with same name and namespace (status != DELETED)
	_, found, err := s.deploymentRepo.GetByNameAndNamespace(ctx, req.Name, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing deployment: %w", err)
	}
	if found {
		// Deployment exists, return conflict error
		return nil, fmt.Errorf(
			"deployment with name '%s' and namespace '%s' already exists",
			req.Name,
			req.Namespace,
		)
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Generate identifier (collision chance is very low)
	identifier, err := utils.GenerateDeploymentIdentifier(req.Name, req.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to generate identifier: %w", err)
	}

	// Convert DTO to model
	deploymentRequest := &models.DeploymentRequest{
		RequestID:   requestID,
		Identifier:  identifier,
		Name:        req.Name,
		Namespace:   req.Namespace,
		RequestType: models.DeploymentRequestTypeCreate,
		Status:      models.DeploymentRequestStatusCreated,
		Image:       req.Image,
		UserID:      userUUID,
		Metadata: models.JSONB{
			"replica_count":  req.Metadata.ReplicaCount,
			"resource_limit": req.Metadata.ResourceLimit,
			"doc_html":       req.Metadata.DocHTML,
		},
	}

	// Save to database via repository
	if err := s.repo.Create(ctx, deploymentRequest); err != nil {
		return nil, fmt.Errorf("failed to create deployment request in database: %w", err)
	}

	// Publish to NATS for worker processing
	if err := s.publisher.Publish(requestID, userID); err != nil {
		s.logger.Error("Failed to publish deployment request to NATS",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to publish deployment request: %w", err)
	}

	s.logger.Info("Deployment request created and published",
		zap.String("request_id", requestID),
		zap.String("identifier", identifier),
		zap.String("name", req.Name),
		zap.String("namespace", req.Namespace),
	)

	return &dto.DeploymentRequestResponse{
		ID:          deploymentRequest.ID,
		RequestID:   deploymentRequest.RequestID,
		Identifier:  deploymentRequest.Identifier,
		Name:        deploymentRequest.Name,
		Namespace:   deploymentRequest.Namespace,
		Image:       deploymentRequest.Image,
		Status:      string(deploymentRequest.Status),
		RequestType: string(deploymentRequest.RequestType),
		Metadata:    map[string]interface{}(deploymentRequest.Metadata),
	}, nil
}

// ListDeploymentRequests returns all deployment requests for the given user
func (s *DeploymentRequestService) ListDeploymentRequests(ctx context.Context, userID string) ([]*dto.DeploymentRequestListResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	requests, err := s.repo.ListByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployment requests: %w", err)
	}

	result := make([]*dto.DeploymentRequestListResponse, 0, len(requests))
	for _, r := range requests {
		result = append(result, &dto.DeploymentRequestListResponse{
			RequestID:     r.RequestID,
			Identifier:    r.Identifier,
			Name:          r.Name,
			Namespace:     r.Namespace,
			Image:         r.Image,
			Status:        string(r.Status),
			RequestType:   string(r.RequestType),
			FailureReason: r.FailureReason,
		})
	}
	return result, nil
}

// GetDeploymentRequest returns the full deployment request by request_id if it belongs to the user
func (s *DeploymentRequestService) GetDeploymentRequest(ctx context.Context, requestID string, userID string) (*dto.DeploymentRequestResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	r, found, err := s.repo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment request: %w", err)
	}
	if !found || r.UserID != userUUID {
		return nil, dto.ErrDeploymentRequestNotFound
	}

	return &dto.DeploymentRequestResponse{
		ID:          r.ID,
		RequestID:   r.RequestID,
		Identifier:  r.Identifier,
		Name:        r.Name,
		Namespace:   r.Namespace,
		Image:       r.Image,
		Status:      string(r.Status),
		RequestType: string(r.RequestType),
		Metadata:    map[string]interface{}(r.Metadata),
	}, nil
}

// UpdateDeploymentRequest handles the business logic for updating a deployment request
func (s *DeploymentRequestService) UpdateDeploymentRequest(
	ctx context.Context,
	identifier string,
	req *dto.UpdateDeploymentRequestMetadata,
	requestID string,
	userID string,
) (*dto.DeploymentRequestResponse, error) {
	s.logger.Info("Updating deployment request",
		zap.String("request_id", requestID),
		zap.String("identifier", identifier),
		zap.String("user_id", userID),
	)

	// Step 1: Check if deployment exists by identifier, is not deleted, and belongs to user
	deployment, found, err := s.deploymentRepo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing deployment: %w", err)
	}
	if !found {
		return nil, dto.ErrDeploymentNotFound
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if deployment belongs to user
	if deployment.UserID != userUUID {
		return nil, dto.ErrDeploymentNotFound
	}

	// Check if deployment is deleted
	if deployment.Status == models.DeploymentStatusDeleted {
		return nil, fmt.Errorf("deployment with identifier '%s' is deleted", identifier)
	}

	// Build metadata map from optional fields
	metadata := make(models.JSONB)
	if req.ReplicaCount != nil {
		metadata["replica_count"] = *req.ReplicaCount
	}
	if req.ResourceLimit != nil {
		metadata["resource_limit"] = req.ResourceLimit
	}
	if req.DocHTML != nil {
		metadata["doc_html"] = *req.DocHTML
	}

	// Convert DTO to model
	deploymentRequest := &models.DeploymentRequest{
		RequestID:   requestID,
		Identifier:  identifier,
		Name:        deployment.Name,
		Namespace:   deployment.Namespace,
		RequestType: models.DeploymentRequestTypeUpdate,
		Status:      models.DeploymentRequestStatusCreated,
		Image:       deployment.Image,
		UserID:      userUUID,
		Metadata:    metadata,
	}

	// Save to database via repository
	if err := s.repo.Create(ctx, deploymentRequest); err != nil {
		return nil, fmt.Errorf("failed to create deployment request in database: %w", err)
	}

	// Publish to NATS for worker processing
	if err := s.publisher.Publish(requestID, userID); err != nil {
		s.logger.Error("Failed to publish deployment request to NATS",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to publish deployment request: %w", err)
	}

	s.logger.Info("Deployment request updated and published",
		zap.String("request_id", requestID),
		zap.String("identifier", identifier),
	)

	return &dto.DeploymentRequestResponse{
		ID:          deploymentRequest.ID,
		RequestID:   deploymentRequest.RequestID,
		Identifier:  deploymentRequest.Identifier,
		Name:        deploymentRequest.Name,
		Namespace:   deploymentRequest.Namespace,
		Image:       deploymentRequest.Image,
		Status:      string(deploymentRequest.Status),
		RequestType: string(deploymentRequest.RequestType),
		Metadata:    map[string]interface{}(deploymentRequest.Metadata),
	}, nil
}

// DeleteDeploymentRequest handles the business logic for deleting a deployment request
func (s *DeploymentRequestService) DeleteDeploymentRequest(
	ctx context.Context,
	identifier string,
	requestID string,
	userID string,
) (*dto.DeploymentRequestResponse, error) {
	s.logger.Info("Deleting deployment request",
		zap.String("request_id", requestID),
		zap.String("identifier", identifier),
		zap.String("user_id", userID),
	)

	// Step 1: Check if deployment exists by identifier, is not deleted, and belongs to user
	deployment, found, err := s.deploymentRepo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing deployment: %w", err)
	}
	if !found {
		return nil, dto.ErrDeploymentNotFound
	}

	// Parse user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Check if deployment belongs to user
	if deployment.UserID != userUUID {
		return nil, dto.ErrDeploymentNotFound
	}

	// Check if deployment is deleted
	if deployment.Status == models.DeploymentStatusDeleted {
		return nil, fmt.Errorf("deployment with identifier '%s' is already deleted", identifier)
	}

	// Convert DTO to model - no metadata needed for delete
	deploymentRequest := &models.DeploymentRequest{
		RequestID:   requestID,
		Identifier:  identifier,
		Name:        deployment.Name,
		Namespace:   deployment.Namespace,
		RequestType: models.DeploymentRequestTypeDelete,
		Status:      models.DeploymentRequestStatusCreated,
		Image:       deployment.Image,
		UserID:      userUUID,
		Metadata:    make(models.JSONB),
	}

	// Save to database via repository
	if err := s.repo.Create(ctx, deploymentRequest); err != nil {
		return nil, fmt.Errorf("failed to create deployment request in database: %w", err)
	}

	// Publish to NATS for worker processing
	if err := s.publisher.Publish(requestID, userID); err != nil {
		s.logger.Error("Failed to publish deployment request to NATS",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to publish deployment request: %w", err)
	}

	s.logger.Info("Deployment request deleted and published",
		zap.String("request_id", requestID),
		zap.String("identifier", identifier),
	)

	return &dto.DeploymentRequestResponse{
		ID:          deploymentRequest.ID,
		RequestID:   deploymentRequest.RequestID,
		Identifier:  deploymentRequest.Identifier,
		Name:        deploymentRequest.Name,
		Namespace:   deploymentRequest.Namespace,
		Image:       deploymentRequest.Image,
		Status:      string(deploymentRequest.Status),
		RequestType: string(deploymentRequest.RequestType),
		Metadata:    map[string]interface{}(deploymentRequest.Metadata),
	}, nil
}
