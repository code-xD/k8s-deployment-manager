package dto

// Example request DTOs - replace with your actual request structures

// CreateDeploymentRequest represents a request to create a deployment
type CreateDeploymentRequest struct {
	Name      string `json:"name" validate:"required,min=3,max=50"`
	Namespace string `json:"namespace" validate:"required,min=1,max=63"`
	Image     string `json:"image" validate:"required"`
	Replicas  int    `json:"replicas" validate:"required,gte=1,lte=100"`
}

// UpdateDeploymentRequest represents a request to update a deployment
type UpdateDeploymentRequest struct {
	Replicas *int    `json:"replicas,omitempty" validate:"omitempty,gte=1,lte=100"`
	Image    *string `json:"image,omitempty" validate:"omitempty"`
}

// CreateDeploymentRequestWithMetadata represents a request to create a deployment with metadata
type CreateDeploymentRequestWithMetadata struct {
	Name      string             `json:"name" validate:"required,min=3,max=50"`
	Namespace string             `json:"namespace" validate:"required,min=1,max=63"`
	Image     string             `json:"image" validate:"required"`
	Metadata  DeploymentMetadata `json:"metadata" validate:"required"`
}

// DeploymentMetadata represents metadata for a deployment
type DeploymentMetadata struct {
	ReplicaCount  int              `json:"replica_count" validate:"required,gte=1,lte=100"`
	ResourceLimit ResourceMetadata `json:"resource_limit" validate:"required"`
	DocHTML       string           `json:"doc_html" validate:"required"`
}

// UpdateDeploymentRequestMetadata represents optional metadata for updating a deployment
// All fields are optional, but if resources is provided, all fields within it must be provided
type UpdateDeploymentRequestMetadata struct {
	ReplicaCount  *int              `json:"replica_count,omitempty" validate:"omitempty,gte=1,lte=100"`
	ResourceLimit *ResourceMetadata `json:"resource_limit,omitempty" validate:"omitempty"`
	DocHTML       *string           `json:"doc_html,omitempty" validate:"omitempty"`
}

type ResourceMetadata struct {
	Request ResourceLimitInfo `json:"request" validate:"required"`
	Limit   ResourceLimitInfo `json:"limit" validate:"required"`
}

// ResourceLimit represents resource limits for a deployment
type ResourceLimitInfo struct {
	CPU    string `json:"cpu" validate:"required"`    // e.g., "500m"
	Memory string `json:"memory" validate:"required"` // e.g., "256Mi"
}

// Example request without body validation (for demonstration)
// Some endpoints might not need body validation
