package processing

import (
    "encoding/json"
    "net/http"
    
    "github.com/google/uuid"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/repository"
)

type Handler struct {
    processingService *ProcessingService
    videoRepo         *repository.VideoRepository
}

func NewHandler(processingService *ProcessingService, videoRepo *repository.VideoRepository) *Handler {
    return &Handler{
        processingService: processingService,
        videoRepo:         videoRepo,
    }
}

// GetProcessingStatus returns the current processing status of a video
func (h *Handler) GetProcessingStatus(w http.ResponseWriter, r *http.Request) {
    videoIDStr := r.URL.Query().Get("video_id")
    if videoIDStr == "" {
        http.Error(w, "video_id parameter required", http.StatusBadRequest)
        return
    }
    
    videoID, err := uuid.Parse(videoIDStr)
    if err != nil {
        http.Error(w, "Invalid video ID format", http.StatusBadRequest)
        return
    }

    video, err := h.videoRepo.GetVideoByVideoID(videoID)
    if err != nil {
        http.Error(w, "Video not found", http.StatusNotFound)
        return
    }
    
    status := map[string]interface{}{
        "video_id":    video.VideoID.String(),
        "title":       video.Title,
        "status":      video.Status,
        "created_at":  video.CreatedAt,
        "updated_at":  video.UpdatedAt,
        "file_size":   video.FileSize,
        "s3_url":      video.S3URL,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}