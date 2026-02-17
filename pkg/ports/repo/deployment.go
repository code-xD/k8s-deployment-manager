package repo

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
)

// Deployment defines the interface for deployment data access
type Deployment interface {
	GetByNameAndNamespace(ctx context.Context, name, namespace string) (*models.Deployment, bool, error)
}
