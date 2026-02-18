package apiService

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
)

// DeploymentRequest defines the interface for creating, listing and getting deployment requests (API stack)
type DeploymentRequest interface {
	CreateDeploymentRequest(ctx context.Context, req *dto.CreateDeploymentRequestWithMetadata, requestID string, userID string) (*dto.DeploymentRequestResponse, error)
	ListDeploymentRequests(ctx context.Context, userID string) ([]*dto.DeploymentRequestListResponse, error)
	GetDeploymentRequest(ctx context.Context, requestID string, userID string) (*dto.DeploymentRequestResponse, error)
}
