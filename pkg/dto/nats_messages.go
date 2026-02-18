package dto

// NATS message header key constants
const (
	HeaderKeyRequestID = "request_id"
	HeaderKeyUserID    = "user_id"
)

// DeploymentRequestMessage is the body for deployment request producer messages
type DeploymentRequestMessage struct {
	RequestID string `json:"request_id"`
}

// DeploymentUpdateMessage is the body for deployment update producer messages
type DeploymentUpdateMessage struct {
	Identifier string `json:"identifier"`
	EventType  string `json:"event_type"` // "add", "update", "delete"
}
