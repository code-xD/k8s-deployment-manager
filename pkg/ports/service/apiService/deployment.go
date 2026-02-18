package apiService

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
)

// Deployment defines the interface for listing and getting deployments (API stack)
type Deployment interface {
	ListDeployments(ctx context.Context, userID string) ([]*dto.DeploymentListResponse, error)
	GetDeployment(ctx context.Context, identifier string, userID string) (*dto.DeploymentResponse, error)
}
