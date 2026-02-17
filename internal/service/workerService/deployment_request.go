package workerService

import (
	"context"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsworker "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/workerService"
	"go.uber.org/zap"
)

// DeploymentRequestService implements worker-side deployment request processing
type DeploymentRequestService struct {
	logger *zap.Logger
}

// NewDeploymentRequestService creates a new worker deployment request service
func NewDeploymentRequestService(logger *zap.Logger) portsworker.DeploymentRequest {
	return &DeploymentRequestService{
		logger: logger,
	}
}

// ProcessDeploymentRequest processes a deployment request message (logs for end-to-end wiring test)
func (s *DeploymentRequestService) ProcessDeploymentRequest(ctx context.Context, msg *dto.DeploymentRequestMessage) error {
	s.logger.Info("Worker service received deployment request message",
		zap.String("request_id", msg.RequestID),
	)
	return nil
}
