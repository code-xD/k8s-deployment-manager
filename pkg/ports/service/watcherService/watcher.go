package watcherService

import "context"

// WatcherService publishes deployment update events (from the informer) to NATS
type WatcherService interface {
	PublishDeploymentUpdate(ctx context.Context, identifier, eventType string) error
}
