package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common contains common fields for all models
type Common struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedOn time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_on"`
	UpdatedOn *time.Time `gorm:"type:timestamp" json:"updated_on"`
}

// BeforeCreate hook to set ID and created timestamp
func (c *Common) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	c.CreatedOn = time.Now()
	return nil
}

// BeforeUpdate hook to update timestamp
func (c *Common) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	c.UpdatedOn = &now
	return nil
}
