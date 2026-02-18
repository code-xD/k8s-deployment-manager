package dto

import "github.com/google/uuid"

// Standard response structures

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a successful response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// DeploymentResponse represents a deployment response
type DeploymentResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Image     string    `json:"image"`
	Replicas  int       `json:"replicas,omitempty"`
	Status    string    `json:"status,omitempty"`
}

// DeploymentEvent represents a deployment event for NATS
type DeploymentEvent struct {
	Type         string    `json:"type"`
	DeploymentID uuid.UUID `json:"deployment_id"`
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace,omitempty"`
}

// DeploymentStatus represents the status of a deployment in Kubernetes
type DeploymentStatus struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	Replicas      int    `json:"replicas"`
	ReadyReplicas int    `json:"ready_replicas"`
	Status        string `json:"status"`
}

// DeploymentRequestResponse represents a deployment request response
type DeploymentRequestResponse struct {
	ID          uuid.UUID              `json:"id"`
	RequestID   string                 `json:"request_id"`
	Identifier  string                 `json:"identifier"`
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace"`
	Image       string                 `json:"image"`
	Status      string                 `json:"status"`
	RequestType string                 `json:"request_type"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DeploymentRequestListResponse represents a deployment request in list responses (no metadata, includes failure_reason)
type DeploymentRequestListResponse struct {
	RequestID     string  `json:"request_id"`
	Identifier    string  `json:"identifier"`
	Name          string  `json:"name"`
	Namespace     string  `json:"namespace"`
	Image         string  `json:"image"`
	Status        string  `json:"status"`
	RequestType   string  `json:"request_type"`
	FailureReason *string `json:"failure_reason,omitempty"`
}
