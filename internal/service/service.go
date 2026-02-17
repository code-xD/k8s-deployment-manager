package service

import "context"

// Add business logic here. Both API and worker call into this layer.
// Implement interfaces from pkg/ports and use them in cmd/api and cmd/worker.

// Example: deployment use-case used by API and worker
type DeploymentService struct{}

func (s *DeploymentService) Process(ctx context.Context, input string) error {
	// Business rules only; no HTTP, no NATS, no DB specifics
	return nil
}
