package streaming

import (
    "context"
    "net/http"
    "github.com/gorilla/mux"
    "encoding/json"
    "log"
)

type Handler struct {
    service *Service
}

func NewHandler(service *Service) *Handler {
    return &Handler{
        service: service,
    }
}

func (h *Handler) GetVideoStreamHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Get video ID from URL params
    vars := mux.Vars(r)
    videoID := vars["videoId"]
    if videoID == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error":   "Video ID is required",
        })
        return
    }

    // Get user ID from context (set by auth middleware)
    userID := r.Context().Value("user_id")
    if userID == nil {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error":   "Authentication required",
        })
        return
    }

    userIDStr := userID.(string)

    // Get streaming URLs from service
    streamingData, err := h.service.GetVideoStreamingURLs(context.Background(), videoID, userIDStr)
    if err != nil {
        log.Printf("Error getting streaming URLs: %v", err)
        
        var statusCode int
        var errorMessage string
        
        switch err.Error() {
        case "video not found":
            statusCode = http.StatusNotFound
            errorMessage = "Video not found"
        case "video not ready":
            statusCode = http.StatusConflict
            errorMessage = "Video is still being processed"
        default:
            statusCode = http.StatusInternalServerError
            errorMessage = "Failed to get streaming URLs"
        }

        w.WriteHeader(statusCode)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error":   errorMessage,
        })
        return
    }

    // Return success response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Streaming URLs retrieved successfully",
        "data":    streamingData,
    })
}