package streaming

import (
    "context"
    "fmt"
    "strings"
    "time"
    "os"
    "gorm.io/gorm"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/entities"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)
type Service struct {
    db         *gorm.DB
    s3Client   *s3.S3
    bucketName string
}

type StreamingData struct {
    VideoID           string          `json:"video_id"`
    Title             string          `json:"title"`
    MasterPlaylistURL string          `json:"master_playlist_url"`
    Qualities         []QualityStream `json:"qualities"`
    Duration          int             `json:"duration,omitempty"`
    ThumbnailURL      string          `json:"thumbnail_url,omitempty"`
}

type QualityStream struct {
    Quality     string `json:"quality"`
    Resolution  string `json:"resolution"`
    Bitrate     string `json:"bitrate"`
    PlaylistURL string `json:"playlist_url"`
}

func NewService(db *gorm.DB) (*Service, error) {
    // Initialize AWS session
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(os.Getenv("AWS_REGION")),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create AWS session: %w", err)
    }

    s3Client := s3.New(sess)
    bucketName := os.Getenv("AWS_S3_BUCKET")
    if bucketName == "" {
        return nil, fmt.Errorf("AWS_S3_BUCKET environment variable is required")
    }

    return &Service{
        db:         db,
        s3Client:   s3Client,
        bucketName: bucketName,
    }, nil
}

func (s *Service) GetVideoStreamingURLs(ctx context.Context, videoID, userID string) (*StreamingData, error) {
    // 1. Get video from database
    var video entities.Video
    err := s.db.Where("id = ? AND user_id = ?", videoID, userID).First(&video).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, fmt.Errorf("video not found")
        }
        return nil, fmt.Errorf("database error: %w", err)
    }

    // 2. Check if video is ready for streaming
    if video.Status != "completed" && video.Status != "ready" {
        return nil, fmt.Errorf("video not ready")
    }

    // 3. Construct S3 paths
    baseS3Path := fmt.Sprintf("adaptive-bitrate-streaming-videos/%s/%s/hls", userID, videoID)

    // 4. Generate master playlist URL
    masterPlaylistKey := fmt.Sprintf("%s/master.m3u8", baseS3Path)
    masterPlaylistURL, err := s.generateSignedURL(masterPlaylistKey, time.Hour*24)
    if err != nil {
        return nil, fmt.Errorf("failed to generate master playlist URL: %w", err)
    }

    // 5. Generate quality-specific URLs
    qualities := []QualityStream{}
    standardQualities := []struct {
        name       string
        resolution string
        bitrate    string
    }{
        {"240p", "426x240", "400k"},
        {"480p", "854x480", "800k"},
        {"720p", "1280x720", "1500k"},
        {"1080p", "1920x1080", "2500k"},
    }

    for _, quality := range standardQualities {
        qualityPlaylistKey := fmt.Sprintf("%s/%s/playlist.m3u8", baseS3Path, quality.name)
        
        // Check if this quality exists
        exists, err := s.checkS3ObjectExists(qualityPlaylistKey)
        if err != nil || !exists {
            continue // Skip this quality if it doesn't exist
        }

        playlistURL, err := s.generateSignedURL(qualityPlaylistKey, time.Hour*24)
        if err != nil {
            continue // Skip this quality if URL generation fails
        }

        qualities = append(qualities, QualityStream{
            Quality:     quality.name,
            Resolution:  quality.resolution,
            Bitrate:     quality.bitrate,
            PlaylistURL: playlistURL,
        })
    }

    // 6. Generate thumbnail URL if exists
    thumbnailURL := ""
    thumbnailKey := fmt.Sprintf("adaptive-bitrate-streaming-videos/%s/%s/thumbnail.jpg", userID, videoID)
    if exists, _ := s.checkS3ObjectExists(thumbnailKey); exists {
        thumbnailURL, _ = s.generateSignedURL(thumbnailKey, time.Hour*24)
    }

    return &StreamingData{
        VideoID:           videoID,
        Title:             video.Title,
        MasterPlaylistURL: masterPlaylistURL,
        Qualities:         qualities,
        Duration:          0, // Add this field to your video entity if needed
        ThumbnailURL:      thumbnailURL,
    }, nil
}

func (s *Service) generateSignedURL(key string, expiration time.Duration) (string, error) {
    req, _ := s.s3Client.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(s.bucketName),
        Key:    aws.String(key),
    })

    url, err := req.Presign(expiration)
    if err != nil {
        return "", fmt.Errorf("failed to generate signed URL: %w", err)
    }

    return url, nil
}

func (s *Service) checkS3ObjectExists(key string) (bool, error) {
    _, err := s.s3Client.HeadObject(&s3.HeadObjectInput{
        Bucket: aws.String(s.bucketName),
        Key:    aws.String(key),
    })

    if err != nil {
        if strings.Contains(err.Error(), "NotFound") {
            return false, nil
        }
        return false, err
    }

    return true, nil
}