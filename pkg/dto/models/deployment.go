package models

import (
	"github.com/google/uuid"
)

// DeploymentStatus represents the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusInitiated DeploymentStatus = "INITIATED"
	DeploymentStatusCreated    DeploymentStatus = "CREATED"
	DeploymentStatusUpdating   DeploymentStatus = "UPDATING"
	DeploymentStatusDeleted    DeploymentStatus = "DELETED"
)

// Deployment represents a deployment
type Deployment struct {
	Common
	Identifier string                 `gorm:"type:varchar(63);uniqueIndex;not null" json:"identifier"`
	Name       string                 `gorm:"type:varchar(255)" json:"name"`
	Namespace  string                 `gorm:"type:varchar(255)" json:"namespace"`
	Image      string                 `gorm:"type:varchar(255)" json:"image"`
	Status     DeploymentStatus        `gorm:"type:varchar(50);not null;index:idx_deployment_user_status" json:"status"`
	UserID     uuid.UUID              `gorm:"type:uuid;not null;index:idx_deployment_user_status" json:"user_id"`
	Metadata   map[string]interface{} `gorm:"type:jsonb" json:"metadata"`
	
	// Foreign key relationship
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName specifies the table name for Deployment
func (Deployment) TableName() string {
	return "deployments"
}
