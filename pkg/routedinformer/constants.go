package routedinformer

// DeploymentEventType represents the type of deployment watch event.
type DeploymentEventType int

const (
	// DeploymentEventAdd is emitted when a deployment is created.
	DeploymentEventAdd DeploymentEventType = iota
	// DeploymentEventUpdate is emitted when a deployment is updated.
	DeploymentEventUpdate
	// DeploymentEventDelete is emitted when a deployment is deleted.
	DeploymentEventDelete
)
