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

// Handle is the DeploymentEventHandler: it extracts namespace, name, and eventType and publishes via WatcherService
func (h *Handler) Handle(ctx context.Context, eventType routedinformer.DeploymentEventType, deployment *appsv1.Deployment) {
	if deployment == nil {
		return
	}
	eventTypeStr := eventType.String()
	if err := h.watcherService.PublishDeploymentUpdate(ctx, deployment.Namespace, deployment.Name, eventTypeStr); err != nil {
		h.logger.Error("Failed to publish deployment update",
			zap.String("namespace", deployment.Namespace),
			zap.String("name", deployment.Name),
			zap.String("event_type", eventTypeStr),
			zap.Error(err),
		)
		return
	}
	h.logger.Debug("Published deployment update",
		zap.String("namespace", deployment.Namespace),
		zap.String("name", deployment.Name),
		zap.String("event_type", eventTypeStr),
	)
}
