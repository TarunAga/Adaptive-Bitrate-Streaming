package entities

import (
    "time"
    "gorm.io/gorm"
)

// Video represents a video in the system
type Video struct {
    ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
    VideoID     string         `gorm:"not null;uniqueIndex;size:36" json:"video_id"` // UUID from S3 upload
    VideoName   string         `gorm:"not null;size:255" json:"video_name"`
    BucketName  string         `gorm:"not null;size:100" json:"bucket_name"`
    BucketKey   string         `gorm:"not null;size:500" json:"bucket_key"` // S3 key path
    UserID      uint           `gorm:"not null;index" json:"user_id"` // Foreign key to users.user_id
    FileSize    int64          `gorm:"not null" json:"file_size"`
    ContentType string         `gorm:"size:100" json:"content_type"`
    S3URL       string         `gorm:"size:1000" json:"s3_url"`
    Status      string         `gorm:"default:'uploaded';size:50" json:"status"` // uploaded, processing, ready, failed
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	User User `gorm:"foreignKey:UserID;references:UserID" json:"-"`
}

// TableName specifies the table name for the Video model
func (Video) TableName() string {
    return "videos"
}