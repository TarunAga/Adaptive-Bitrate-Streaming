package repository

import (
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/entities"
    "gorm.io/gorm"
	"github.com/google/uuid"
)

type VideoRepository struct {
    db *gorm.DB
}

func NewVideoRepository(db *gorm.DB) *VideoRepository {
    return &VideoRepository{db: db}
}

// CreateVideo creates a new video record
func (r *VideoRepository) CreateVideo(video *entities.Video) (*entities.Video, error) {
    err := r.db.Create(video).Error
    if err != nil {
        return nil, err
    }
    return video, nil
}

// GetVideoByVideoID finds video by video_id (UUID)
func (r *VideoRepository) GetVideoByVideoID(videoID uuid.UUID) (*entities.Video, error) {
    var video entities.Video
    result := r.db.Where("video_id = ?", videoID).First(&video)
    
    if result.Error != nil {
        return nil, result.Error
    }
    
    return &video, nil
}

// GetVideosByUserID gets all videos for a user
func (r *VideoRepository) GetVideosByUserID(userID uuid.UUID) ([]*entities.Video, error) {
    var videos []*entities.Video
    result := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&videos)
    
    if result.Error != nil {
        return nil, result.Error
    }
    
    return videos, nil
}

// UpdateVideoStatus updates the status of a video
func (r *VideoRepository) UpdateVideoStatus(videoID uuid.UUID, status string) error {
    result := r.db.Model(&entities.Video{}).Where("video_id = ?", videoID).Update("status", status)
    return result.Error
}

// GetAllVideos gets all videos with user information
func (r *VideoRepository) GetAllVideos() ([]entities.Video, error) {
    var videos []entities.Video
    result := r.db.Preload("User").Find(&videos)
    
    if result.Error != nil {
        return nil, result.Error
    }
    
    return videos, nil
}
// ✅ ADD: UpdateVideo method
func (r *VideoRepository) UpdateVideo(video *entities.Video) error {
    return r.db.Save(video).Error
}

// ✅ ADD: GetDB method (for processing service)
func (r *VideoRepository) GetDB() *gorm.DB {
    return r.db
}