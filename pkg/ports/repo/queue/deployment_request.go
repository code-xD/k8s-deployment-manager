package queue

// DeploymentRequest publishes deployment request messages to NATS
type DeploymentRequest interface {
	Publish(requestID, userID string) error
}
