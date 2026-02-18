package apiService

import (
	"context"
	"fmt"
	"time"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	portsapi "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/apiService"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeploymentService implements the deployment business logic for the API
type DeploymentService struct {
	deploymentRepo portsdb.Deployment
	logger         *zap.Logger
}

// NewDeploymentService creates a new DeploymentService with injected dependencies
func NewDeploymentService(
	deploymentRepo portsdb.Deployment,
	logger *zap.Logger,
) portsapi.Deployment {
	return &DeploymentService{
		deploymentRepo: deploymentRepo,
		logger:         logger,
	}
}

// ListDeployments returns all deployments for the given user
func (s *DeploymentService) ListDeployments(ctx context.Context, userID string) ([]*dto.DeploymentListResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	deployments, err := s.deploymentRepo.ListByUserID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	result := make([]*dto.DeploymentListResponse, 0, len(deployments))
	for _, d := range deployments {
		updatedAt := ""
		if d.UpdatedOn != nil {
			updatedAt = d.UpdatedOn.Format(time.RFC3339)
		}

		result = append(result, &dto.DeploymentListResponse{
			Identifier: d.Identifier,
			CreatedAt:  d.CreatedOn.Format(time.RFC3339),
			UpdatedAt:  updatedAt,
			Status:     string(d.Status),
			Name:       d.Name,
			Namespace:  d.Namespace,
		})
	}
	return result, nil
}

// GetDeployment returns the full deployment by identifier if it belongs to the user
func (s *DeploymentService) GetDeployment(ctx context.Context, identifier string, userID string) (*dto.DeploymentResponse, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	d, found, err := s.deploymentRepo.GetByIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}
	if !found || d.UserID != userUUID {
		return nil, dto.ErrDeploymentNotFound
	}

	updatedAt := ""
	if d.UpdatedOn != nil {
		updatedAt = d.UpdatedOn.Format(time.RFC3339)
	}

	return &dto.DeploymentResponse{
		ID:         d.ID,
		Identifier: d.Identifier,
		Name:       d.Name,
		Namespace:  d.Namespace,
		Image:      d.Image,
		Status:     string(d.Status),
		CreatedAt:  d.CreatedOn.Format(time.RFC3339),
		UpdatedAt:  updatedAt,
		Metadata:   map[string]interface{}(d.Metadata),
	}, nil
}
