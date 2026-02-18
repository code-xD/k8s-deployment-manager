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

// String returns the event type as a string for serialization (e.g. "add", "update", "delete").
func (e DeploymentEventType) String() string {
	switch e {
	case DeploymentEventAdd:
		return "add"
	case DeploymentEventUpdate:
		return "update"
	case DeploymentEventDelete:
		return "delete"
	default:
		return "unknown"
	}
}
