package processing
import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "strconv"
    "strings"
    "time"
    "path/filepath"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/google/uuid"
)

type VideoMetadata struct {
    Width    int     `json:"width"`
    Height   int     `json:"height"`
    Duration float64 `json:"duration"`
    Bitrate  string  `json:"bitrate"`
    Format   string  `json:"format"`
}

func (ps *ProcessingService) downloadFromS3(bucketName, s3Key, localPath string) (int64, error) {
    result, err := ps.s3Client.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(s3Key),
    })
    if err != nil {
        return 0, fmt.Errorf("failed to get object from S3: %w", err)
    }
    defer result.Body.Close()
    
    file, err := os.Create(localPath)
    if err != nil {
        return 0, fmt.Errorf("failed to create local file: %w", err)
    }
    defer file.Close()
    
    size, err := io.Copy(file, result.Body)
    if err != nil {
        return 0, fmt.Errorf("failed to download file: %w", err)
    }
    
    return size, nil
}
func (ps *ProcessingService) getVideoMetadata(videoPath string) (*VideoMetadata, error) {
    cmd := exec.Command("ffprobe",
        "-v", "quiet",
        "-print_format", "json",
        "-show_format",
        "-show_streams",
        videoPath,
    )
    
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("ffprobe failed: %w", err)
    }
    
    var probeResult struct {
        Streams []struct {
            CodecType string `json:"codec_type"`
            Width     int    `json:"width"`
            Height    int    `json:"height"`
        } `json:"streams"`
        Format struct {
            Duration string `json:"duration"`
            Bitrate  string `json:"bit_rate"`
        } `json:"format"`
    }
    
    err = json.Unmarshal(output, &probeResult)
    if err != nil {
        return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
    }
    
    // Find video stream
    var width, height int
    for _, stream := range probeResult.Streams {
        if stream.CodecType == "video" {
            width = stream.Width
            height = stream.Height
            break
        }
    }
    
    duration, _ := strconv.ParseFloat(probeResult.Format.Duration, 64)
    
    return &VideoMetadata{
        Width:    width,
        Height:   height,
        Duration: duration,
        Bitrate:  probeResult.Format.Bitrate,
        Format:   "mp4",
    }, nil
}
func (ps *ProcessingService) defineQualityTargets(metadata *VideoMetadata) []VideoQuality {
    // Base quality definitions
    baseQualities := []struct {
        name    string
        height  int
        bitrate string    // ✅ FIXED: Keep as string
    }{
        {"240p", 240, "500k"},
        {"480p", 480, "1500k"},
        {"720p", 720, "3000k"},
        {"1080p", 1080, "5000k"},
    }
    
    var qualities []VideoQuality
    aspectRatio := float64(metadata.Width) / float64(metadata.Height)
    
    // Add qualities that are smaller than or equal to original
    for _, base := range baseQualities {
        if base.height <= metadata.Height {
            width := int(float64(base.height) * aspectRatio)
            // Ensure even numbers for video encoding
            if width%2 != 0 {
                width--
            }
            if base.height%2 != 0 {
                base.height--
            }
            
            qualities = append(qualities, VideoQuality{
                Name:    base.name,
                Height:  base.height,
                Width:   width,
                Bitrate: base.bitrate,    // ✅ FIXED: String value
            })
        }
    }
    
    // Always add original quality if it's not already included
    originalIncluded := false
    for _, q := range qualities {
        if q.Height == metadata.Height {
            originalIncluded = true
            break
        }
    }
    
    if !originalIncluded {
        qualities = append(qualities, VideoQuality{
            Name:    "original",
            Height:  metadata.Height,
            Width:   metadata.Width,
            Bitrate: "8000k",    // ✅ FIXED: String value
        })
    }
    
    return qualities
}
func (ps *ProcessingService) generateHLSStreamWorker(ctx context.Context, inputPath, tempDir string, quality VideoQuality, userID, videoID string) (VideoQuality, error) {
    qualityDir := filepath.Join(tempDir, quality.Name)
    err := os.MkdirAll(qualityDir, 0755)
    if err != nil {
        return VideoQuality{}, fmt.Errorf("failed to create quality directory: %w", err)
    }

    playlistPath := filepath.Join(qualityDir, "playlist.m3u8")
    segmentPath := filepath.Join(qualityDir, "segment_%03d.ts")

    cmd := exec.CommandContext(ctx, "ffmpeg",
        "-i", inputPath,
        "-vf", fmt.Sprintf("scale=%d:-2", quality.Width),
        "-c:a", "aac", "-ar", "48000", "-b:a", "128k",
        "-c:v", "h264", "-profile:v", "main", "-crf", "20",
        "-g", "48", "-keyint_min", "48", "-sc_threshold", "0",
        "-b:v", quality.Bitrate,
        "-maxrate", fmt.Sprintf("%dk", parseInt(quality.Bitrate)*2),
        "-bufsize", fmt.Sprintf("%dk", parseInt(quality.Bitrate)*4),
        "-hls_time", "6", "-hls_playlist_type", "vod",
        "-hls_segment_filename", segmentPath,
        "-f", "hls", playlistPath,
    )

    output, err := cmd.CombinedOutput()
    if err != nil {
        return VideoQuality{}, fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
    }

    // Calculate output metrics
    totalSize, segmentCount := ps.calculateHLSOutput(qualityDir)
    
    // ✅ FIX: Upload to S3 with correct parameters
    s3KeyPrefix := fmt.Sprintf("adaptive-bitrate-streaming-videos/%s/%s/hls/%s", userID, videoID, quality.Name)
    playlistURL, err := ps.uploadHLSToS3(qualityDir, s3KeyPrefix, uuid.MustParse(videoID))
    if err != nil {
        return VideoQuality{}, err
    }

    return VideoQuality{
        Name:         quality.Name,
        Height:       quality.Height,
        Width:        quality.Width,
        Bitrate:      quality.Bitrate,
        Size:         totalSize,
        S3KeyPrefix:  s3KeyPrefix,
        PlaylistURL:  playlistURL,
        SegmentCount: segmentCount,
    }, nil
}
func (ps *ProcessingService) uploadHLSToS3(qualityDir, s3KeyPrefix string, videoID uuid.UUID) (string, error) {
    files, err := os.ReadDir(qualityDir)
    if err != nil {
        return "", err
    }
    
    var playlistURL string
    
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        
        filePath := filepath.Join(qualityDir, file.Name())
        s3Key := fmt.Sprintf("%s/%s", s3KeyPrefix, file.Name())
        
        contentType := "application/octet-stream"
        if strings.HasSuffix(file.Name(), ".m3u8") {
            contentType = "application/x-mpegURL"
        } else if strings.HasSuffix(file.Name(), ".ts") {
            contentType = "video/MP2T"
        }
        
        err = ps.uploadFileToS3(filePath, s3Key, contentType)
        if err != nil {
            return "", fmt.Errorf("failed to upload %s: %w", file.Name(), err)
        }
        
        if file.Name() == "playlist.m3u8" {
            playlistURL = fmt.Sprintf("https://adaptive-bitrate-streaming-videos.s3.ap-south-1.amazonaws.com/%s", s3Key)
        }
    }
    
    return playlistURL, nil
}
func (ps *ProcessingService) generateMasterPlaylist(hlsDir string, qualities []VideoQuality, videoID uuid.UUID) (string, error) {
    // Extract userID from one of the quality S3 key prefixes
    if len(qualities) == 0 {
        return "", fmt.Errorf("no qualities to generate master playlist")
    }
    
    // Parse userID from S3 key prefix: adaptive-bitrate-streaming-videos/{userID}/{videoID}/hls/{quality}
    parts := strings.Split(qualities[0].S3KeyPrefix, "/")
    if len(parts) < 3 {
        return "", fmt.Errorf("invalid S3 key prefix format")
    }
    userID := parts[1]
    
    // Create master playlist content
    content := "#EXTM3U\n#EXT-X-VERSION:3\n\n"
    
    for _, quality := range qualities {
        bandwidth := parseInt(quality.Bitrate) * 1000
        content += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n", 
            bandwidth, quality.Width, quality.Height)
        content += fmt.Sprintf("%s/playlist.m3u8\n\n", quality.Name)
    }
    
    // Create temp file for master playlist
    tempDir := filepath.Join(ps.tempDir, "master_"+videoID.String())
    os.MkdirAll(tempDir, 0755)
    defer os.RemoveAll(tempDir)
    
    masterPath := filepath.Join(tempDir, "master.m3u8")
    err := os.WriteFile(masterPath, []byte(content), 0644)
    if err != nil {
        return "", err
    }
    
    // Upload master playlist to S3
    s3Key := fmt.Sprintf("adaptive-bitrate-streaming-videos/%s/%s/hls/master.m3u8", userID, videoID.String())
    err = ps.uploadFileToS3(masterPath, s3Key, "application/x-mpegURL")
    if err != nil {
        return "", err
    }
    
    masterURL := fmt.Sprintf("https://adaptive-bitrate-streaming-videos.s3.ap-south-1.amazonaws.com/%s", s3Key)
    return masterURL, nil
}
func (ps *ProcessingService) updateVideoRecord(videoID uuid.UUID, qualities []VideoQuality) error {
    video, err := ps.videoRepo.GetVideoByVideoID(videoID)
    if err != nil {
        return fmt.Errorf("failed to get video record: %w", err)
    }
    
    // Update video status
    video.Status = "completed"
    
    // You might want to store qualities in a separate table or as JSON
    // For now, we'll update the main video record
    err = ps.videoRepo.UpdateVideo(video)
    if err != nil {
        return fmt.Errorf("failed to update video record: %w", err)
    }
    
    return nil
}
func parseInt(s string) int {
    s = strings.TrimSuffix(s, "k")
    s = strings.TrimSuffix(s, "K")
    val, err := strconv.Atoi(s)
    if err != nil {
        return 1000 // Default value if parsing fails
    }
    return val
}
func (ps *ProcessingService) uploadFileToS3(filePath, s3Key, contentType string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()
    
    _, err = ps.s3Client.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(ps.bucketName),
        Key:         aws.String(s3Key),
        Body:        file,
        ContentType: aws.String(contentType),
        Metadata: map[string]*string{
            "processed-at": aws.String(time.Now().Format(time.RFC3339)),
            "processor":    aws.String("adaptive-streaming-service"),
        },
    })
    
    if err != nil {
        return fmt.Errorf("failed to upload to S3: %w", err)
    }
    
    return nil
}
// func (ps *ProcessingService) generateMasterPlaylist(hlsDir string, qualities []VideoQuality, videoID uuid.UUID) (string, error) {
//     // Extract userID from one of the quality S3 key prefixes
//     if len(qualities) == 0 {
//         return "", fmt.Errorf("no qualities to generate master playlist")
//     }
    
//     // Parse userID from S3 key prefix: adaptive-bitrate-streaming-videos/{userID}/{videoID}/hls/{quality}
//     parts := strings.Split(qualities[0].S3KeyPrefix, "/")
//     if len(parts) < 3 {
//         return "", fmt.Errorf("invalid S3 key prefix format")
//     }
//     userID := parts[1]
    
//     // Create master playlist content
//     content := "#EXTM3U\n#EXT-X-VERSION:3\n\n"
    
//     for _, quality := range qualities {
//         bandwidth := parseInt(quality.Bitrate) * 1000
//         content += fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n", 
//             bandwidth, quality.Width, quality.Height)
//         content += fmt.Sprintf("%s/playlist.m3u8\n\n", quality.Name)
//     }
    
//     // Create temp file for master playlist
//     tempDir := filepath.Join(ps.tempDir, "master_"+videoID.String())
//     os.MkdirAll(tempDir, 0755)
//     defer os.RemoveAll(tempDir)
    
//     masterPath := filepath.Join(tempDir, "master.m3u8")
//     err := os.WriteFile(masterPath, []byte(content), 0644)
//     if err != nil {
//         return "", err
//     }
    
//     // Upload master playlist to S3
//     s3Key := fmt.Sprintf("adaptive-bitrate-streaming-videos/%s/%s/hls/master.m3u8", userID, videoID.String())
//     err = ps.uploadFileToS3(masterPath, s3Key, "application/x-mpegURL")
//     if err != nil {
//         return "", err
//     }
    
//     masterURL := fmt.Sprintf("https://adaptive-bitrate-streaming-videos.s3.ap-south-1.amazonaws.com/%s", s3Key)
//     return masterURL, nil
// }

func (ps *ProcessingService) calculateHLSOutput(dir string) (int64, int) {
    files, err := os.ReadDir(dir)
    if err != nil {
        return 0, 0
    }
    
    var totalSize int64
    segmentCount := 0
    
    for _, file := range files {
        if !file.IsDir() {
            info, err := file.Info()
            if err != nil {
                continue
            }
            totalSize += info.Size()
            
            if strings.HasSuffix(file.Name(), ".ts") {
                segmentCount++
            }
        }
    }
    
    return totalSize, segmentCount
}
