package upload

import (
	"fmt"
	"log"
	"mime/multipart"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// BucketName is the S3 bucket where videos will be uploaded
	BucketName = "adaptive-bitrate-streaming-videos"
	
	// AWS Region - change this to your preferred region
	AWSRegion = "us-east-1"
)

// Service handles video upload operations
type Service struct {
	s3Client *s3.S3
}

// NewService creates a new upload service instance
func NewService() (*Service, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AWSRegion),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &Service{
		s3Client: s3.New(sess),
	}, nil
}

// UploadRequest represents the upload request data
type UploadRequest struct {
	UserID string
	Title  string
	File   multipart.File
	Header *multipart.FileHeader
}

// UploadResponse represents the upload response
type UploadResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	S3Key     string `json:"s3_key,omitempty"`
	S3URL     string `json:"s3_url,omitempty"`
	FileSize  int64  `json:"file_size,omitempty"`
}

// UploadVideo uploads a video file to S3
func (s *Service) UploadVideo(req *UploadRequest) (*UploadResponse, error) {
	// Generate S3 key
	s3Key := generateS3Key(req.UserID, req.Title, req.Header.Filename)
	
	// Validate file size (optional - you can set limits)
	if req.Header.Size > 500*1024*1024 { // 500MB limit
		return &UploadResponse{
			Success: false,
			Message: "File size exceeds 500MB limit",
		}, nil
	}

	// Upload to S3
	_, err := s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(BucketName),
		Key:           aws.String(s3Key),
		Body:          req.File,
		ContentType:   aws.String(req.Header.Header.Get("Content-Type")),
		ContentLength: aws.Int64(req.Header.Size),
		Metadata: map[string]*string{
			"user-id":       aws.String(req.UserID),
			"original-name": aws.String(req.Header.Filename),
			"title":         aws.String(req.Title),
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

	return &UploadResponse{
		Success:  true,
		Message:  "Video uploaded successfully",
		S3Key:    s3Key,
		S3URL:    s3URL,
		FileSize: req.Header.Size,
	}, nil
}

// generateS3Key creates a unique S3 key for the uploaded video
func generateS3Key(userID, title, filename string) string {
	// Clean the title to make it filesystem-safe
	cleanTitle := strings.ReplaceAll(title, " ", "_")
	cleanTitle = strings.ReplaceAll(cleanTitle, "/", "_")
	cleanTitle = strings.ReplaceAll(cleanTitle, "\\", "_")
	
	// Get file extension
	parts := strings.Split(filename, ".")
	ext := ""
	if len(parts) > 1 {
		ext = "." + parts[len(parts)-1]
	}
	
	// Format: userId/UploadedVideo_TITLE_timestamp.ext
	return fmt.Sprintf("%s/UploadedVideo_%s%s", userID, cleanTitle, ext)
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
