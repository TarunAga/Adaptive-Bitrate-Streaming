package upload

import (
    "context"
    "fmt"
    "log"
    "mime/multipart"
    "path/filepath"
    "strings"
    "time"              

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/entities"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/repository"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/processing"  
)

const (
    // BucketName is the S3 bucket where videos will be uploaded
    BucketName = "adaptive-bitrate-streaming-videos"
    
    // AWS Region - change this to your preferred region
    AWSRegion = "ap-south-1"
)

// Service handles video upload operations
type Service struct {
    s3Client        *s3.S3
    userRepo        *repository.UserRepository
    videoRepo       *repository.VideoRepository
}

// NewService creates a new upload service instance
func NewService(db *gorm.DB) (*Service, error) {
    // Create AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(AWSRegion),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create AWS session: %w", err)
    }

    return &Service{
        s3Client:  s3.New(sess),
        userRepo:  repository.NewUserRepository(db),
        videoRepo: repository.NewVideoRepository(db),
    }, nil
}

// UploadRequest represents the upload request data
type UploadRequest struct {
    UserId  uuid.UUID
    Title   string
    File    multipart.File
    Header  *multipart.FileHeader
}

// UploadResponse represents the upload response
type UploadResponse struct {
    Success   bool      `json:"success"`
    Message   string    `json:"message"`
    VideoID   string    `json:"video_id"`           
    UserID    string    `json:"user_id,omitempty"`
    Title     string    `json:"title,omitempty"`
    Filename  string    `json:"filename,omitempty"`
    S3Key     string    `json:"s3_key,omitempty"`
    S3URL     string    `json:"s3_url,omitempty"`
    FileSize  int64     `json:"file_size,omitempty"`
    Duration  float64   `json:"duration,omitempty"`
    Status    string    `json:"status,omitempty"`
    CreatedAt time.Time `json:"created_at,omitempty"`
}
type ErrorResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Error   string `json:"error,omitempty"`
}
// UploadVideo uploads a video file to S3 and saves metadata to database
func (s *Service) UploadVideo(req *UploadRequest) (*UploadResponse, error) {
    // Generate unique video ID (UUID)
    videoID := uuid.New().String()
    // Get user by ID
    user, err := s.userRepo.GetUserByID(req.UserId)
    if err != nil {
        return &UploadResponse{
            Success: false,
            Message: "Failed to process user information",
        }, fmt.Errorf("user operation failed: %w", err)
    }
    filename := generateUniqueFilename(req.Header.Filename)
    // Generate S3 key with new structure
    s3Key := fmt.Sprintf("adaptive-bitrate-streaming-videos/%s/%s/original/%s", 
        user.UserID.String(), 
        videoID,
        filename)

    log.Printf("Generated S3 key: %s for user: %s (ID: %d)", s3Key, user.UserName, user.UserID)
    
    // Validate file size (optional - you can set limits)
    if req.Header.Size > 500*1024*1024 { // 500MB limit
        return &UploadResponse{
            Success: false,
            Message: "File size exceeds 500MB limit",
        }, nil
    }

    // Upload to S3
    _, err = s.s3Client.PutObject(&s3.PutObjectInput{
        Bucket:        aws.String(BucketName),
        Key:           aws.String(s3Key),
        Body:          req.File,
        ContentType:   aws.String(req.Header.Header.Get("Content-Type")),
        ContentLength: aws.Int64(req.Header.Size),
        Metadata: map[string]*string{
            "user-id":       aws.String(user.UserID.String()),
            "user-name":     aws.String(user.UserName),
            "original-name": aws.String(req.Header.Filename),
            "title":         aws.String(req.Title),
            "video-id":      aws.String(videoID),
        },
    })

    if err != nil {
        log.Printf("Failed to upload to S3: %v", err)
        return &UploadResponse{
            Success: false,
            Message: "Failed to upload video to S3",
        }, fmt.Errorf("S3 upload failed: %w", err)
    }

    // Generate S3 URL
    s3URL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", BucketName, AWSRegion, s3Key)

    video := &entities.Video{
        VideoID:  uuid.MustParse(videoID),
        UserID:   user.UserID,          
        Title:    req.Title,            
        Filename: filename,              
        S3Key:    s3Key,               
        S3URL:    s3URL,                 
        FileSize: req.Header.Size,       
        Status:   "uploaded",           
    }

    createdVideo, err := s.videoRepo.CreateVideo(video)
    if err != nil {
        log.Printf("Failed to save video metadata: %v", err)
        // Note: S3 upload succeeded, but database save failed
        // You might want to implement cleanup logic here
        return &UploadResponse{
            Success: false,
            Message: "Video uploaded to S3 but failed to save metadata",
        }, fmt.Errorf("database save failed: %w", err)
    }
    go s.processVideoBackground(createdVideo.UserID, BucketName, s3Key, createdVideo.VideoID)

    return &UploadResponse{
        Success:   true,
        Message:   "Video uploaded successfully",
        VideoID:   createdVideo.VideoID.String(),
        UserID:    user.UserID.String(),
        Title:     createdVideo.Title,
        Filename:  createdVideo.Filename,
        S3Key:     createdVideo.S3Key,
        S3URL:     createdVideo.S3URL,
        FileSize:  createdVideo.FileSize,
        Status:    createdVideo.Status,
        CreatedAt: createdVideo.CreatedAt,
    }, nil
}
func generateUniqueFilename(originalFilename string) string {
    ext := filepath.Ext(originalFilename)
    name := strings.TrimSuffix(originalFilename, ext)
    timestamp := time.Now().Format("20060102_150405")
    uuidStr := uuid.New().String()[:8]
    return fmt.Sprintf("%s_%s_%s%s", name, timestamp, uuidStr, ext)
}

// generateS3Key creates a unique S3 key for the uploaded video
// Structure: userId/videoTitle_randomuuid/filename.ext
func generateS3Key(userId uuid.UUID, title, videoID, filename string) string {
    // Clean the title to make it filesystem-safe
    cleanTitle := strings.ReplaceAll(title, " ", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, "/", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, "\\", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, ":", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, "?", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, "*", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, "<", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, ">", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, "|", "_")
    cleanTitle = strings.ReplaceAll(cleanTitle, "\"", "_")


    // Get file extension
    parts := strings.Split(filename, ".")
    ext := ""
    if len(parts) > 1 {
        ext = "." + parts[len(parts)-1]
    }

    // Structure: userId/videoTitle_randomuuid/filename.ext
    // Example: john_doe/My_Video_Title_a1b2c3d4-e5f6-7890-abcd-ef1234567890/My_Video_Title.mp4
    folderName := fmt.Sprintf("%s_%s", cleanTitle, videoID)
    return fmt.Sprintf("%s/%s/%s%s", userId.String(), folderName, cleanTitle, ext)
}

// CreateBucket creates the S3 bucket if it doesn't exist (optional utility function)
func (s *Service) CreateBucket() error {
    _, err := s.s3Client.CreateBucket(&s3.CreateBucketInput{
        Bucket: aws.String(BucketName),
    })
    
    if err != nil {
        // Check if bucket already exists
        if strings.Contains(err.Error(), "BucketAlreadyExists") ||
           strings.Contains(err.Error(), "BucketAlreadyOwnedByYou") {
            return nil // Bucket already exists, which is fine
        }
        return fmt.Errorf("failed to create bucket: %w", err)
    }
    
    return nil
}
func (s *Service) processVideoBackground(userID uuid.UUID, bucketName, s3Key string, videoID uuid.UUID) {
    log.Printf("Starting background processing for video: %s", videoID.String())
    
    // Update status to processing
    video, err := s.videoRepo.GetVideoByVideoID(videoID)
    if err != nil {
        log.Printf("Failed to get video for processing: %v", err)
        return
    }
    
    video.Status = "processing"
    s.videoRepo.UpdateVideo(video)
    
    // Create processing service
processingService := processing.NewProcessingService(
    s.s3Client,    // *s3.S3
    s.videoRepo,   // *repository.VideoRepository  
    bucketName,    // string (bucket name)
)
    // Process video
    result, err := processingService.ProcessVideo(context.Background(), bucketName, s3Key, videoID)
    if err != nil {
        log.Printf("Video processing failed for %s: %v", videoID.String(), err)
        
        // Update status to failed
        video.Status = "failed"
        s.videoRepo.UpdateVideo(video)
        return
    }
    
    log.Printf("Video processing completed for %s: %d qualities generated", 
        videoID.String(), len(result.Qualities))
}

// GetUserVideos retrieves all videos for a specific user
func (s *Service) GetUserVideos(userID uuid.UUID) ([]*entities.Video, error) {
    // Get user by ID to ensure they exist
    user, err := s.userRepo.GetUserByID(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    if user == nil {
        return nil, fmt.Errorf("user not found")
    }
    
    // Get all videos for the user
    videos, err := s.videoRepo.GetVideosByUserID(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user videos: %w", err)
    }
    
    return videos, nil
}