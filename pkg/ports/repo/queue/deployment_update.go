package queue

// DeploymentUpdate publishes deployment update messages to NATS
type DeploymentUpdate interface {
	Publish(identifier, eventType string) error
}
