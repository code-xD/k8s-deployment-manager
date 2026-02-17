package db

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
)

// DeploymentRequest defines the interface for deployment request data access
type DeploymentRequest interface {
	Create(ctx context.Context, deployment *models.DeploymentRequest) error
	GetByIdentifier(ctx context.Context, identifier string) (*models.DeploymentRequest, error)
	GetByRequestID(ctx context.Context, requestID string) (*models.DeploymentRequest, bool, error)
	UpdateStatus(ctx context.Context, id interface{}, status models.DeploymentRequestStatus) error
}
