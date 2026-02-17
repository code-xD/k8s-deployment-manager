package service

import (
	"context"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	portsqueue "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/queue"
	portsservice "github.com/code-xd/k8s-deployment-manager/pkg/ports/service"
	"github.com/code-xd/k8s-deployment-manager/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeploymentRequestService implements the deployment request business logic
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
) portsservice.DeploymentRequest {
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
