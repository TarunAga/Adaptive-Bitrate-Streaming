package entities

import (
    "time"
    "gorm.io/gorm"
    "github.com/google/uuid"
)

// Video represents a video in the system
type Video struct {
    VideoID     uuid.UUID      `gorm:"type:uuid;primaryKey" json:"video_id"`
    UserID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"` // ✅ Changed to UUID
    Title       string         `gorm:"not null;size:255" json:"title"`
    Filename    string         `gorm:"not null;size:255" json:"filename"`
    S3Key       string         `gorm:"not null;size:500" json:"s3_key"`
    S3URL       string         `gorm:"not null;size:500" json:"s3_url"`
    FileSize    int64          `gorm:"not null" json:"file_size"`
    Duration    float64        `json:"duration,omitempty"`
    Status      string         `gorm:"default:'uploaded'" json:"status"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

    // Relationship
    User User `gorm:"foreignKey:UserID;references:UserID" json:"user,omitempty"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (v *Video) BeforeCreate(tx *gorm.DB) error {
    if v.VideoID == uuid.Nil {
        v.VideoID = uuid.New()
    }
    return nil
}

// TableName specifies the table name for the Video model
func (Video) TableName() string {
    return "videos"
}