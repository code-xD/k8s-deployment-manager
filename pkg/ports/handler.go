package ports

import "github.com/code-xd/k8s-deployment-manager/pkg/dto"

// Handler defines the interface for all handlers that can register routes
type Handler interface {
	GetRoutes() []dto.RouteDefinition
}
