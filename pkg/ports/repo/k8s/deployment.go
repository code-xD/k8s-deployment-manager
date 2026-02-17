package k8s

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	appsv1 "k8s.io/api/apps/v1"
)

// Deployment defines the interface for Kubernetes deployment operations
type Deployment interface {
	Create(ctx context.Context, req *models.DeploymentRequest) (*appsv1.Deployment, error)
}
