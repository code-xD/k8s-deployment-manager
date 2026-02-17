package postgres

import (
	"context"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/internal/database/query"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
)

// DeploymentRequestRepository implements the deployment request repository interface
type DeploymentRequestRepository struct {
	db *common.DB
}

// NewDeploymentRequestRepository creates a new deployment request repository
func NewDeploymentRequestRepository(db *common.DB) portsdb.DeploymentRequest {
	return &DeploymentRequestRepository{
		db: db,
	}
}

// Create creates a new deployment request in the database
func (r *DeploymentRequestRepository) Create(ctx context.Context, deployment *models.DeploymentRequest) error {
	q := query.Use(r.db.DB)
	return q.DeploymentRequest.WithContext(ctx).Create(deployment)
}

// GetByIdentifier retrieves a deployment request by identifier
// Returns error if deployment exists and is in CREATED or SUCCESS status
func (r *DeploymentRequestRepository) GetByIdentifier(ctx context.Context, identifier string) (*models.DeploymentRequest, error) {
	q := query.Use(r.db.DB)
	deployment, err := q.DeploymentRequest.WithContext(ctx).
		Where(q.DeploymentRequest.Identifier.Eq(identifier)).
		Where(q.DeploymentRequest.Status.In(
			string(models.DeploymentRequestStatusCreated),
			string(models.DeploymentRequestStatusSuccess),
		)).
		First()
	if err != nil {
		return nil, fmt.Errorf("deployment not found: %w", err)
	}
	return deployment, nil
}

// GetByRequestID retrieves a deployment request by request ID
// Returns single object (at most one), boolean indicating if found, and error
func (r *DeploymentRequestRepository) GetByRequestID(ctx context.Context, requestID string) (*models.DeploymentRequest, bool, error) {
	q := query.Use(r.db.DB)
	deployments, err := q.DeploymentRequest.WithContext(ctx).
		Where(q.DeploymentRequest.RequestID.Eq(requestID)).
		Find()
	if err != nil {
		return nil, false, fmt.Errorf("failed to query deployment request: %w", err)
	}
	if len(deployments) > 0 {
		return deployments[0], true, nil
	}
	return nil, false, nil
}
