package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/internal/database/query"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	"gorm.io/gorm"
)

// DeploymentRepository implements the deployment repository interface
type DeploymentRepository struct {
	db *common.DB
}

// NewDeploymentRepository creates a new deployment repository
func NewDeploymentRepository(db *common.DB) portsdb.Deployment {
	return &DeploymentRepository{
		db: db,
	}
}

// GetByNameAndNamespace retrieves a deployment by name and namespace where status is not DELETED
// Returns single object (at most one), boolean indicating if found, and error
func (r *DeploymentRepository) GetByNameAndNamespace(ctx context.Context, name, namespace string) (*models.Deployment, bool, error) {
	q := query.Use(r.db.DB)
	deployments, err := q.Deployment.WithContext(ctx).
		Where(q.Deployment.Name.Eq(name), q.Deployment.Namespace.Eq(namespace)).
		Where(q.Deployment.Status.Neq(string(models.DeploymentStatusDeleted))).
		Find()
	if err != nil {
		return nil, false, fmt.Errorf("failed to query deployment: %w", err)
	}
	if len(deployments) > 0 {
		return deployments[0], true, nil
	}
	return nil, false, nil
}

// Upsert creates or updates a deployment based on identifier unique constraint.
// Only updates if the new resourceVersion is different from the current one.
func (r *DeploymentRepository) Upsert(ctx context.Context, deployment *models.Deployment) error {
	q := query.Use(r.db.DB)
	
	// First, try to find existing deployment by identifier
	existing, err := q.Deployment.WithContext(ctx).
		Where(q.Deployment.Identifier.Eq(deployment.Identifier)).
		First()
	
	if err != nil {
		// If not found, create new deployment
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := q.Deployment.WithContext(ctx).Create(deployment); err != nil {
				return fmt.Errorf("failed to create deployment: %w", err)
			}
			return nil
		}
		// Other error occurred
		return fmt.Errorf("failed to query deployment: %w", err)
	}
	
	// If found, check if resourceVersion is different
	if existing.ResourceVersion == deployment.ResourceVersion {
		// ResourceVersion is the same, skip update
		return nil
	}
	
	// Update existing deployment with new values
	deployment.ID = existing.ID
	deployment.CreatedOn = existing.CreatedOn
	
	if _, err := q.Deployment.WithContext(ctx).
		Where(q.Deployment.ID.Eq(existing.ID)).
		Updates(deployment); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}
	
	return nil
}
