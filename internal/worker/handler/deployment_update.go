package handler

import (
	"context"
	"encoding/json"

	"github.com/code-xd/k8s-deployment-manager/pkg/consumer"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	portsworker "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/workerService"
	"go.uber.org/zap"
)

// DeploymentUpdateHandler handles deployment update messages from the NATS queue
type DeploymentUpdateHandler struct {
	deploymentUpdate portsworker.DeploymentUpdate
	log              *zap.Logger
}

// NewDeploymentUpdateHandler creates a new deployment update handler
func NewDeploymentUpdateHandler(deploymentUpdate portsworker.DeploymentUpdate, log *zap.Logger) *DeploymentUpdateHandler {
	return &DeploymentUpdateHandler{
		deploymentUpdate: deploymentUpdate,
		log:              log,
	}
}

// Handle processes a deployment update message: logs headers, unmarshals body, passes to service
func (h *DeploymentUpdateHandler) Handle(ctx context.Context, msg *consumer.Message) error {
	headers := consumer.HeadersFromContext(ctx)
	if headers != nil {
		for k, v := range headers {
			if len(v) > 0 {
				h.log.Info("Message header", zap.String("key", k), zap.String("value", v[0]))
			}
		}
	}

	var body dto.DeploymentUpdateMessage
	if err := json.Unmarshal(msg.Data, &body); err != nil {
		h.log.Error("Failed to unmarshal deployment update message", zap.Error(err))
		return err
	}

	return h.deploymentUpdate.ProcessDeploymentUpdate(ctx, &body)
}
