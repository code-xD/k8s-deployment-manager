package postgres

import (
	"context"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/internal/database/query"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
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
