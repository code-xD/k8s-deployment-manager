package watcher

import (
	"context"

	portswatcher "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/watcherService"
	"github.com/code-xd/k8s-deployment-manager/pkg/routedinformer"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
)

// Handler handles deployment events from the informer and publishes updates via WatcherService
type Handler struct {
	watcherService portswatcher.WatcherService
	logger         *zap.Logger
}

// NewHandler creates a new deployment event handler
func NewHandler(watcherService portswatcher.WatcherService, logger *zap.Logger) *Handler {
	return &Handler{
		watcherService: watcherService,
		logger:         logger,
	}
}

// Handle is the DeploymentEventHandler: it builds identifier and eventType and publishes via WatcherService
func (h *Handler) Handle(ctx context.Context, eventType routedinformer.DeploymentEventType, deployment *appsv1.Deployment) {
	if deployment == nil {
		return
	}
	identifier := deployment.Namespace + "/" + deployment.Name
	eventTypeStr := eventType.String()
	if err := h.watcherService.PublishDeploymentUpdate(ctx, identifier, eventTypeStr); err != nil {
		h.logger.Error("Failed to publish deployment update",
			zap.String("identifier", identifier),
			zap.String("event_type", eventTypeStr),
			zap.Error(err),
		)
		return
	}
	h.logger.Debug("Published deployment update",
		zap.String("identifier", identifier),
		zap.String("event_type", eventTypeStr),
	)
}
