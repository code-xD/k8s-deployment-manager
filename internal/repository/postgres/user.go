package postgres

import (
	"context"
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/internal/database/query"
	"github.com/code-xd/k8s-deployment-manager/internal/repository/postgres/common"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	portsdb "github.com/code-xd/k8s-deployment-manager/pkg/ports/repo/db"
)

// UserRepository implements the user repository interface
type UserRepository struct {
	db *common.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *common.DB) portsdb.User {
	return &UserRepository{
		db: db,
	}
}

// GetByExternalID retrieves a user by external ID
func (r *UserRepository) GetByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	q := query.Use(r.db.DB)
	user, err := q.User.WithContext(ctx).
		Where(q.User.UserExternalID.Eq(externalID)).
		First()
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	q := query.Use(r.db.DB)
	return q.User.WithContext(ctx).Create(user)
}
