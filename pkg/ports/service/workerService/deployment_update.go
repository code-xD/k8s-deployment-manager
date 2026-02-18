package workerService

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
)

// DeploymentUpdate defines the interface for processing deployment update messages (worker stack)
type DeploymentUpdate interface {
	ProcessDeploymentUpdate(ctx context.Context, msg *dto.DeploymentUpdateMessage) error
}
