package upload

import (
    "fmt"
    "log"
    "mime/multipart"
    "strings"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/entities"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/repository"
)

const (
    // BucketName is the S3 bucket where videos will be uploaded
    BucketName = ""
    
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
    UserName string // Changed from UserID to UserName for input
    Title    string
    File     multipart.File
    Header   *multipart.FileHeader
}

// UploadResponse represents the upload response
type UploadResponse struct {
    Success     bool   `json:"success"`
    Message     string `json:"message"`
    S3Key       string `json:"s3_key,omitempty"`
    S3URL       string `json:"s3_url,omitempty"`
    FileSize    int64  `json:"file_size,omitempty"`
    VideoID     string `json:"video_id,omitempty"`
    UserID      uint   `json:"user_id,omitempty"`
    DatabaseID  uint   `json:"database_id,omitempty"` // Auto-generated database ID
}

// UploadVideo uploads a video file to S3 and saves metadata to database
func (s *Service) UploadVideo(req *UploadRequest) (*UploadResponse, error) {
    // Generate unique video ID (UUID)
    videoID := uuid.New().String()
    // Get user by username
    user, err := s.userRepo.GetUserByUserName(req.UserName)
    if err != nil {
        return &UploadResponse{
            Success: false,
            Message: "Failed to process user information",
        }, fmt.Errorf("user operation failed: %w", err)
    }
    
    // Generate S3 key with new structure
    s3Key := generateS3Key(req.UserName, req.Title, videoID, req.Header.Filename)
    
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
            "user-id":       aws.String(fmt.Sprintf("%d", user.UserID)),
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

    // Save video metadata to database
    video := &entities.Video{
        VideoID:     videoID,
        VideoName:   req.Title,
        BucketName:  BucketName,
        BucketKey:   s3Key,
        UserID:      user.UserID,
        FileSize:    req.Header.Size,
        ContentType: req.Header.Header.Get("Content-Type"),
        S3URL:       s3URL,
        Status:      "uploaded",
    }

    err = s.videoRepo.CreateVideo(video)
    if err != nil {
        log.Printf("Failed to save video metadata: %v", err)
        // Note: S3 upload succeeded, but database save failed
        // You might want to implement cleanup logic here
        return &UploadResponse{
            Success: false,
            Message: "Video uploaded to S3 but failed to save metadata",
        }, fmt.Errorf("database save failed: %w", err)
    }

    log.Printf("Video uploaded successfully - DB ID: %d, Video ID: %s, User: %s", 
        video.ID, videoID, user.UserName)

    return &UploadResponse{
        Success:    true,
        Message:    "Video uploaded successfully",
        S3Key:      s3Key,
        S3URL:      s3URL,
        FileSize:   req.Header.Size,
        VideoID:    videoID,
        UserID:     user.UserID,
        DatabaseID: video.ID,
    }, nil
}

// generateS3Key creates a unique S3 key for the uploaded video
// Structure: username/videoTitle_randomuuid/filename.ext
func generateS3Key(userName, title, videoID, filename string) string {
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
    
    // Clean the userName as well
    cleanUserName := strings.ReplaceAll(userName, "/", "_")
    cleanUserName = strings.ReplaceAll(cleanUserName, "\\", "_")
    
    // Get file extension
    parts := strings.Split(filename, ".")
    ext := ""
    if len(parts) > 1 {
        ext = "." + parts[len(parts)-1]
    }
    
    // Structure: username/videoTitle_randomuuid/filename.ext
    // Example: john_doe/My_Video_Title_a1b2c3d4-e5f6-7890-abcd-ef1234567890/My_Video_Title.mp4
    folderName := fmt.Sprintf("%s_%s", cleanTitle, videoID)
    return fmt.Sprintf("%s/%s/%s%s", cleanUserName, folderName, cleanTitle, ext)
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