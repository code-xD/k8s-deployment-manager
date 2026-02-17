package repo

// DeploymentRequestPublisher publishes deployment request messages to NATS
type DeploymentRequestPublisher interface {
	Publish(requestID, userID string) error
}
