package watcherService

import (
	"context"
	"fmt"

	portsqueue "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/queue"
	portswatcher "github.com/code-xd/k8s-deployment-manager/pkg/ports/service/watcherService"
	"go.uber.org/zap"
)

// WatcherService publishes deployment update events to NATS
type WatcherService struct {
	deploymentUpdate portsqueue.DeploymentUpdate
	logger           *zap.Logger
}

// NewWatcherService creates a new watcher service
func NewWatcherService(
	deploymentUpdate portsqueue.DeploymentUpdate,
	logger *zap.Logger,
) portswatcher.WatcherService {
	return &WatcherService{
		deploymentUpdate: deploymentUpdate,
		logger:           logger,
	}
}

// PublishDeploymentUpdate publishes a deployment update message to NATS
func (s *WatcherService) PublishDeploymentUpdate(ctx context.Context, identifier, eventType string) error {
	if err := s.deploymentUpdate.Publish(identifier, eventType); err != nil {
		return fmt.Errorf("publish deployment update: %w", err)
	}
	return nil
}
