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
	Replicas int `json:"replicas" validate:"required,gte=1,lte=100"`
}

// Example request without body validation (for demonstration)
// Some endpoints might not need body validation
