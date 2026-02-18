package db

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	"github.com/google/uuid"
)

// Deployment defines the interface for deployment data access
type Deployment interface {
	GetByNameAndNamespace(ctx context.Context, name, namespace string) (*models.Deployment, bool, error)
	GetByIdentifier(ctx context.Context, identifier string) (*models.Deployment, bool, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Deployment, error)
	Upsert(ctx context.Context, deployment *models.Deployment) error
	Update(ctx context.Context, deployment *models.Deployment) error
}
