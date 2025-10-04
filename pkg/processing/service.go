package processing
import (
    "context"
    "fmt" 
    "log"
    "os"
    "path/filepath"
    "strings" 
    "time"
    "github.com/aws/aws-sdk-go/service/s3"
    "github.com/google/uuid"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/repository"
)

type ProcessingService struct {
	s3Client *s3.S3
	videoRepo *repository.VideoRepository
	bucketName string
	tempDir string
    workerPool *WorkerPool
}

type VideoQuality struct {
    Name         string  `json:"name"`         // "240p", "480p", "720p"
    Height       int     `json:"height"`       
    Width        int     `json:"width"`        
    Bitrate      string  `json:"bitrate"`      
    Size         int64   `json:"size"`         // Total size of all segments
    S3KeyPrefix  string  `json:"s3_key_prefix"`// S3 prefix for this quality
    PlaylistURL  string  `json:"playlist_url"` // URL to quality playlist
    SegmentCount int     `json:"segment_count"`// Number of .ts segments
    Duration     float64 `json:"duration"`     
}

type ProcessingResult struct {
    VideoID          uuid.UUID       `json:"video_id"`
    Status           string          `json:"status"`
    Message          string          `json:"message"`
    OriginalSize     int64           `json:"original_size"`
    ProcessedAt      time.Time       `json:"processed_at"`
    ProcessingTime   time.Duration   `json:"processing_time"`
    Qualities        []VideoQuality  `json:"qualities"`
    TotalSize        int64           `json:"total_size"`
    CompressionRatio float64         `json:"compression_ratio"`
    MasterPlaylistURL string         `json:"master_playlist_url"`
    StreamingType     string         `json:"streaming_type"`
}
func NewProcessingService(s3Client *s3.S3, videoRepo *repository.VideoRepository, bucketName string) *ProcessingService {
    // Auto-detect temp directory based on OS
    var tempDir string
    if os.Getenv("OS") == "Windows_NT" {
        tempDir = "C:\\temp\\video-processing"
    } else {
        tempDir = "/tmp/video-processing"
    }
    
    // Create the directory if it doesn't exist
    err := os.MkdirAll(tempDir, 0755)
    if err != nil {
        log.Printf("Warning: Failed to create temp directory %s: %v", tempDir, err)
    }
    
    ps := &ProcessingService{
        s3Client:   s3Client,
        videoRepo:  videoRepo,
        bucketName: bucketName,
        tempDir:    tempDir,
    }
    
    ps.workerPool = NewWorkerPool(4, ps) // 4 workers for 4 qualities
    ps.workerPool.Start()
    
    return ps
}
func (ps *ProcessingService) ProcessVideo(ctx context.Context, bucketName, s3Key string, videoID uuid.UUID) (*ProcessingResult, error) {
    startTime := time.Now()
    
    log.Printf("🎬 Starting PARALLEL HLS processing for VideoID: %s", videoID.String())
    
    // Extract userID from s3Key
    parts := strings.Split(s3Key, "/")
    if len(parts) < 4 {
        return nil, fmt.Errorf("invalid S3 key format: %s", s3Key)
    }
    userID := parts[1]
    
    // Step 1: Download and prepare
    processDir := filepath.Join(ps.tempDir, "download_"+videoID.String())
    originalPath := filepath.Join(processDir, "original.mp4")
    os.MkdirAll(processDir, 0755)
    defer os.RemoveAll(processDir)
    
    originalSize, err := ps.downloadFromS3(bucketName, s3Key, originalPath)
    if err != nil {
        return nil, fmt.Errorf("failed to download video: %w", err)
    }
    
    metadata, err := ps.getVideoMetadata(originalPath)
    if err != nil {
        return nil, fmt.Errorf("failed to get metadata: %w", err)
    }
    
    qualities := ps.defineQualityTargets(metadata)
    
    // ✅ NEW: Submit all jobs to worker pool
    log.Printf("🚀 Submitting %d jobs to worker pool for parallel processing", len(qualities))
    
    expectedJobs := len(qualities)
    
    // Submit all jobs simultaneously
    for i, quality := range qualities {
        jobID := uuid.New()
        job := ProcessingJob{
            JobID:       jobID,
            VideoID:     videoID,
            UserID:      userID,
            InputPath:   originalPath,
            Quality:     quality,
            S3KeyPrefix: fmt.Sprintf("adaptive-bitrate-streaming-videos/%s/%s/hls/%s", 
                userID, videoID.String(), quality.Name),
            Priority:    i + 1,
        }
        
        ps.workerPool.SubmitJob(job)
    }
    
    // ✅ NEW: Collect results as they complete
    var processedQualities []VideoQuality
    var totalProcessedSize int64
    completedJobs := 0
    
    timeout := time.After(15 * time.Minute) // Generous timeout
    
    for completedJobs < expectedJobs {
        select {
        case result := <-ps.workerPool.GetResults():
            completedJobs++
            
            if result.Success {
                processedQualities = append(processedQualities, result.Quality)
                totalProcessedSize += result.OutputSize
                log.Printf("✅ %s completed in %v (%d/%d)", 
                    result.Quality.Name, result.ProcessingTime, completedJobs, expectedJobs)
            } else {
                log.Printf("❌ %s failed: %s (%d/%d)", 
                    result.Quality.Name, result.Error, completedJobs, expectedJobs)
            }
            
        case <-timeout:
            log.Printf("⏰ Timeout! Completed %d/%d", completedJobs, expectedJobs)
            break
            
        case <-ctx.Done():
            log.Printf("🛑 Cancelled! Completed %d/%d", completedJobs, expectedJobs)
            return nil, fmt.Errorf("processing cancelled")
        }
    }
    
    // Generate master playlist
    masterPlaylistURL, err := ps.generateMasterPlaylist("", processedQualities, videoID)
    if err != nil {
        log.Printf("Failed to create master playlist: %v", err)
    }
    
    processingTime := time.Since(startTime)
    
    result := &ProcessingResult{
        VideoID:           videoID,
        Status:            "completed",
        Message:           fmt.Sprintf("Parallel processing completed: %d/%d qualities", len(processedQualities), expectedJobs),
        OriginalSize:      originalSize,
        ProcessedAt:       time.Now(),
        ProcessingTime:    processingTime,
        Qualities:         processedQualities,
        TotalSize:         totalProcessedSize,
        CompressionRatio:  float64(totalProcessedSize) / float64(originalSize),
        MasterPlaylistURL: masterPlaylistURL,
        StreamingType:     "HLS",
    }
    
    log.Printf("🎉 PARALLEL processing completed in %v!", processingTime)
    
    return result, nil
}

func (ps *ProcessingService) Shutdown() {
    if ps.workerPool != nil {
        ps.workerPool.Stop()
    }
}