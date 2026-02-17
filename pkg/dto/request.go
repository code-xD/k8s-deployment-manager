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
	Name      string                 `json:"name" validate:"required,min=3,max=50"`
	Namespace string                 `json:"namespace" validate:"required,min=1,max=63"`
	Image     string                 `json:"image" validate:"required"`
	Metadata  DeploymentMetadata     `json:"metadata" validate:"required"`
}

// DeploymentMetadata represents metadata for a deployment
type DeploymentMetadata struct {
	ReplicaCount int    `json:"replica_count" validate:"required,gte=1,lte=100"`
	ResourceLimit string `json:"resource_limit" validate:"required"`
	DocHTML      string `json:"doc_html" validate:"required"`
}

// Example request without body validation (for demonstration)
// Some endpoints might not need body validation
