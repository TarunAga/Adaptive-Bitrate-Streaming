package entities

import (
    "time"
    "gorm.io/gorm"
)

// User represents a user in the system
type User struct {
    UserID    uint           `gorm:"primaryKey;autoIncrement" json:"user_id"`
    UserName  string         `gorm:"not null;uniqueIndex;size:100" json:"user_name"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	PasswordHash string         `gorm:"not null;size:255" json:"-"`
	Email 	  string         `gorm:"not null;uniqueIndex;size:100" json:"email"`
	FirstName string         `gorm:"size:100" json:"first_name,omitempty"`
	LastName  string         `gorm:"size:100" json:"last_name,omitempty"`
	IsActive bool           `gorm:"default:true" json:"is_active"`
	

	Videos []Video `gorm:"foreignKey:UserID;references:UserID;constraint:OnDelete:CASCADE" json:"videos,omitempty"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
    return "users"
}