package upload

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for video uploads
type Handler struct {
	service *Service
}

// NewHandler creates a new upload handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// UploadVideoHandler handles multipart video upload requests
func (h *Handler) UploadVideoHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "X-User-ID header is required")
		return
	}
	userId, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid X-User-ID header")
		return
	}

	// Parse multipart form (32MB max memory)
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Printf("Failed to parse multipart form: %v", err)
		respondWithError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}
	title := r.FormValue("title")

	if userId == uuid.Nil {
		respondWithError(w, http.StatusBadRequest, "userId is required")
		return
	}

	if title == "" {
		respondWithError(w, http.StatusBadRequest, "title is required")
		return
	}

	// Get the file from form
	file, fileHeader, err := r.FormFile("video")
	if err != nil {
		log.Printf("Failed to get file from form: %v", err)
		respondWithError(w, http.StatusBadRequest, "Failed to get video file")
		return
	}
	defer file.Close()

	// Validate file type (basic validation)
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Check if it's a video file
	if !isVideoFile(contentType, fileHeader.Filename) {
		respondWithError(w, http.StatusBadRequest, "Only video files are allowed")
		return
	}

	// Create upload request
    uploadReq := &UploadRequest{
        UserId: userId,
        Title:  title,
        File:   file,
        Header: fileHeader,
    }

	// Upload to S3
	response, err := h.service.UploadVideo(uploadReq)
	if err != nil {
		log.Printf("Upload failed: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Upload failed")
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if response.Success {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	
	json.NewEncoder(w).Encode(response)
}

// HealthCheckHandler provides a health check endpoint
func (h *Handler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "upload-service",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUploadInfoHandler provides information about upload limits and requirements
func (h *Handler) GetUploadInfoHandler(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"max_file_size_mb": 500,
		"allowed_formats":  []string{"mp4", "avi", "mov", "mkv", "webm"},
		"bucket_name":      BucketName,
		"required_fields":  []string{"userName", "title", "video"},
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}

// GetUserVideosHandler returns all videos for the authenticated user
func (h *Handler) GetUserVideosHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "X-User-ID header is required")
		return
	}
	
	userId, err := uuid.Parse(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid X-User-ID header")
		return
	}

	// Get user videos from service
	videos, err := h.service.GetUserVideos(userId)
	if err != nil {
		log.Printf("Failed to get user videos: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to get videos")
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Videos retrieved successfully",
		"videos":  videos,
		"count":   len(videos),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// isVideoFile checks if the uploaded file is a video file
func isVideoFile(contentType, filename string) bool {
	// Check content type
	videoTypes := []string{
		"video/mp4",
		"video/avi",
		"video/quicktime",
		"video/x-msvideo",
		"video/x-matroska",
		"video/webm",
	}

	for _, vType := range videoTypes {
		if contentType == vType {
			return true
		}
	}

	// Check file extension as fallback
	videoExtensions := []string{".mp4", ".avi", ".mov", ".mkv", ".webm"}
	for _, ext := range videoExtensions {
		if len(filename) > len(ext) && 
		   filename[len(filename)-len(ext):] == ext {
			return true
		}
	}

	return false
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	response := UploadResponse{
		Success: false,
		Message: message,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
