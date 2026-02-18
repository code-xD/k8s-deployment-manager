package k8s

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentManager defines the interface for Kubernetes deployment operations
type DeploymentManager interface {
	Create(ctx context.Context, req *models.DeploymentRequest) (*appsv1.Deployment, error)
	Get(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	// GetOptional returns the deployment if found; second return is false if the deployment does not exist in the cluster.
	GetOptional(ctx context.Context, namespace, name string) (*appsv1.Deployment, bool, error)
	Update(ctx context.Context, req *models.DeploymentRequest, existingDeployment *appsv1.Deployment) (*appsv1.Deployment, error)
	Delete(ctx context.Context, namespace, name string) error
}
