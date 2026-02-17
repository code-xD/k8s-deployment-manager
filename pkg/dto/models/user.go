package models

// User represents a user in the system
type User struct {
	Common
	UserExternalID string `gorm:"uniqueIndex;not null" json:"user_external_id"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}
