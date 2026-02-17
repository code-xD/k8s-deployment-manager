package models

import (
	"github.com/google/uuid"
)

// DeploymentRequestStatus represents the status of a deployment request
type DeploymentRequestStatus string

const (
	DeploymentRequestStatusCreated DeploymentRequestStatus = "CREATED"
	DeploymentRequestStatusSuccess DeploymentRequestStatus = "SUCCESS"
	DeploymentRequestStatusFailure DeploymentRequestStatus = "FAILURE"
)

// DeploymentRequestType represents the type of deployment request
type DeploymentRequestType string

const (
	DeploymentRequestTypeCreate DeploymentRequestType = "CREATE"
	DeploymentRequestTypeUpdate DeploymentRequestType = "UPDATE"
	DeploymentRequestTypeDelete DeploymentRequestType = "DELETE"
)

// DeploymentRequest represents a deployment request
type DeploymentRequest struct {
	Common
	RequestID   string                  `gorm:"uniqueIndex;not null" json:"request_id"`
	Identifier  string                  `gorm:"type:varchar(63);not null" json:"identifier"`
	Name        string                  `gorm:"type:varchar(255);index:idx_deployment_request_name_namespace,priority:2" json:"name"`
	Namespace   string                  `gorm:"type:varchar(255);index:idx_deployment_request_name_namespace,priority:1" json:"namespace"`
	RequestType DeploymentRequestType   `gorm:"type:varchar(50);not null" json:"request_type"`
	UserID      uuid.UUID               `gorm:"type:uuid;not null;index:idx_deployment_request_user_status" json:"user_id"`
	Status      DeploymentRequestStatus `gorm:"type:varchar(50);not null;index:idx_deployment_request_user_status" json:"status"`
	Image       string                  `gorm:"type:varchar(255);not null" json:"image"`
	Metadata    map[string]interface{}  `gorm:"type:jsonb" json:"metadata"`

	// Foreign key relationship
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName specifies the table name for DeploymentRequest
func (DeploymentRequest) TableName() string {
	return "deployment_requests"
}
