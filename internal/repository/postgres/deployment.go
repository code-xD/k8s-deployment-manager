package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/internal/database/query"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
	"github.com/google/uuid"
	"go.uber.org/zap"
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

// GetByIdentifier retrieves a deployment by identifier.
// Returns (deployment, true, nil) if found, (nil, false, nil) if not found.
func (r *DeploymentRepository) GetByIdentifier(ctx context.Context, identifier string) (*models.Deployment, bool, error) {
	q := query.Use(r.db.DB)
	existing, err := q.Deployment.WithContext(ctx).
		Where(q.Deployment.Identifier.Eq(identifier)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to query deployment: %w", err)
	}
	return existing, true, nil
}

// ListByUserID retrieves all deployments for a given user ID
func (r *DeploymentRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Deployment, error) {
	q := query.Use(r.db.DB)
	deployments, err := q.Deployment.WithContext(ctx).
		Where(q.Deployment.UserID.Eq(userID)).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to query deployments by user ID: %w", err)
	}
	return deployments, nil
}

// Update updates an existing deployment by ID (e.g. status and UpdatedOn).
func (r *DeploymentRepository) Update(ctx context.Context, deployment *models.Deployment) error {
	q := query.Use(r.db.DB)
	_, err := q.Deployment.WithContext(ctx).
		Where(q.Deployment.ID.Eq(deployment.ID)).
		Updates(deployment)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}
	return nil
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
			dto.Log.Info("Deployment not found, creating new deployment",
				zap.String("identifier", deployment.Identifier),
			)
			err = q.Deployment.WithContext(ctx).Create(deployment)
		}
		return err
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
