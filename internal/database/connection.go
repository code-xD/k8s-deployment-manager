package database

import (
	"fmt"

	"github.com/code-xd/k8s-deployment-manager/pkg/dto"
	"github.com/code-xd/k8s-deployment-manager/pkg/dto/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB holds the database connection
type DB struct {
	*gorm.DB
	logger *zap.Logger
}

// NewDB creates a new database connection
func NewDB(cfg *dto.DatabaseConfig, log *zap.Logger) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	dbInstance := &DB{
		DB:     db,
		logger: log,
	}

	// Run migrations
	if err := dbInstance.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbInstance, nil
}

// Migrate runs database migrations
func (db *DB) Migrate() error {
	db.logger.Info("Running database migrations...")
	
	err := db.AutoMigrate(
		&models.User{},
		&models.DeploymentRequest{},
		&models.Deployment{},
	)
	
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	db.logger.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
