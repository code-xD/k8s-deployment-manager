package workerService

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
)

// DeploymentRequest defines the interface for processing deployment request messages (worker stack)
type DeploymentRequest interface {
	ProcessDeploymentRequest(ctx context.Context, msg *dto.DeploymentRequestMessage) error
}
