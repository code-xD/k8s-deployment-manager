package handler

import (
	"context"
	"encoding/json"

	"github.com/code-xd/k8s-deployment-manager/pkg/consumer"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsworker "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/workerService"
	"go.uber.org/zap"
)

// DeploymentRequestHandler handles deployment request messages from the NATS queue
type DeploymentRequestHandler struct {
	deploymentRequest portsworker.DeploymentRequest
	log     *zap.Logger
}

// NewDeploymentRequestHandler creates a new deployment request handler
func NewDeploymentRequestHandler(deploymentRequest portsworker.DeploymentRequest, log *zap.Logger) *DeploymentRequestHandler {
	return &DeploymentRequestHandler{
		deploymentRequest: deploymentRequest,
		log:     log,
	}
}

// Handle processes a deployment request message: logs headers, unmarshals body, passes to service
func (h *DeploymentRequestHandler) Handle(ctx context.Context, msg *consumer.Message) error {
	headers := consumer.HeadersFromContext(ctx)
	if headers != nil {
		for k, v := range headers {
			if len(v) > 0 {
				h.log.Info("Message header", zap.String("key", k), zap.String("value", v[0]))
			}
		}
	}

	var body dto.DeploymentRequestMessage
	if err := json.Unmarshal(msg.Data, &body); err != nil {
		h.log.Error("Failed to unmarshal deployment request message", zap.Error(err))
		return err
	}

	return h.deploymentRequest.ProcessDeploymentRequest(ctx, &body)
}
