package repo

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
)

// User defines the interface for user data access
type User interface {
	GetByExternalID(ctx context.Context, externalID string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
}
