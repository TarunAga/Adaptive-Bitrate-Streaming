package processing

import (
    "context"
    "fmt"
    "log"
    "os"
    "sync"
    "time"
    
    "github.com/google/uuid"
)

// WorkerPool manages concurrent video processing
type WorkerPool struct {
    workerCount       int
    jobQueue          chan ProcessingJob
    resultQueue       chan WorkerResult
    workers           []Worker
    ctx               context.Context
    cancel            context.CancelFunc
    wg                sync.WaitGroup
    processingService *ProcessingService
}

// ProcessingJob represents a single quality processing task
type ProcessingJob struct {
    JobID       uuid.UUID     `json:"job_id"`
    VideoID     uuid.UUID     `json:"video_id"`
    UserID      string        `json:"user_id"`
    InputPath   string        `json:"input_path"`
    Quality     VideoQuality  `json:"quality"`
    S3KeyPrefix string        `json:"s3_key_prefix"`
    Priority    int           `json:"priority"`
}

// WorkerResult represents the result of a processing job
type WorkerResult struct {
    JobID           uuid.UUID     `json:"job_id"`
    VideoID         uuid.UUID     `json:"video_id"`
    Quality         VideoQuality  `json:"quality"`
    Success         bool          `json:"success"`
    Error           string        `json:"error,omitempty"`
    ProcessingTime  time.Duration `json:"processing_time"`
    SegmentCount    int           `json:"segment_count"`
    OutputSize      int64         `json:"output_size"`
}

// Worker represents a single worker goroutine
type Worker struct {
    ID                int
    jobQueue          <-chan ProcessingJob
    resultQueue       chan<- WorkerResult
    processingService *ProcessingService
    ctx               context.Context
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workerCount int, processingService *ProcessingService) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())
    
    return &WorkerPool{
        workerCount:       workerCount,
        jobQueue:          make(chan ProcessingJob, workerCount*2),
        resultQueue:       make(chan WorkerResult, workerCount*2),
        ctx:               ctx,
        cancel:            cancel,
        processingService: processingService,
    }
}

// Start initializes and starts all workers
func (wp *WorkerPool) Start() {
    log.Printf("🚀 Starting worker pool with %d workers", wp.workerCount)
    
    wp.workers = make([]Worker, wp.workerCount)
    
    for i := 0; i < wp.workerCount; i++ {
        worker := Worker{
            ID:                i + 1,
            jobQueue:          wp.jobQueue,
            resultQueue:       wp.resultQueue,
            processingService: wp.processingService,
            ctx:               wp.ctx,
        }
        
        wp.workers[i] = worker
        wp.wg.Add(1)
        
        go worker.start(&wp.wg)
        log.Printf("✅ Worker %d started", worker.ID)
    }
}

// SubmitJob submits a new processing job
func (wp *WorkerPool) SubmitJob(job ProcessingJob) {
    select {
    case wp.jobQueue <- job:
        log.Printf("📝 Job submitted: %s for quality %s", job.JobID.String(), job.Quality.Name)
    case <-wp.ctx.Done():
        log.Printf("❌ Cannot submit job - worker pool is shutting down")
    default:
        log.Printf("⚠️ Job queue is full, trying with timeout...")
        select {
        case wp.jobQueue <- job:
            log.Printf("📝 Job submitted after wait: %s", job.JobID.String())
        case <-time.After(10 * time.Second):
            log.Printf("❌ Job submission timeout: %s", job.JobID.String())
        }
    }
}

// GetResults returns the result channel
func (wp *WorkerPool) GetResults() <-chan WorkerResult {
    return wp.resultQueue
}

// Stop gracefully stops the worker pool
func (wp *WorkerPool) Stop() {
    log.Printf("🛑 Stopping worker pool...")
    wp.cancel()
    close(wp.jobQueue)
    wp.wg.Wait()
    close(wp.resultQueue)
    log.Printf("✅ Worker pool stopped")
}

// Worker start method
func (w *Worker) start(wg *sync.WaitGroup) {
    defer wg.Done()
    
    for {
        select {
        case job, ok := <-w.jobQueue:
            if !ok {
                log.Printf("🔴 Worker %d: Queue closed, shutting down", w.ID)
                return
            }
            
            log.Printf("⚡ Worker %d processing %s", w.ID, job.Quality.Name)
            result := w.processJob(job)
            
            select {
            case w.resultQueue <- result:
                status := "✅"
                if !result.Success {
                    status = "❌"
                }
                log.Printf("%s Worker %d completed %s in %v", 
                    status, w.ID, job.Quality.Name, result.ProcessingTime)
            case <-w.ctx.Done():
                return
            }
            
        case <-w.ctx.Done():
            log.Printf("🔴 Worker %d: Context cancelled", w.ID)
            return
        }
    }
}

// processJob processes a single job
func (w *Worker) processJob(job ProcessingJob) WorkerResult {
    startTime := time.Now()
    
    result := WorkerResult{
        JobID:   job.JobID,
        VideoID: job.VideoID,
        Quality: job.Quality,
        Success: false,
    }
    
    // Create temp directory for this specific job
    tempDir := fmt.Sprintf("/tmp/worker_%d_%s", w.ID, job.JobID.String())
    if os.Getenv("OS") == "Windows_NT" {
        tempDir = fmt.Sprintf("C:\\temp\\worker_%d_%s", w.ID, job.JobID.String())
    }
    defer os.RemoveAll(tempDir)
    
    // Process the quality
    processedQuality, err := w.processingService.generateHLSStreamWorker(
        w.ctx, 
        job.InputPath, 
        tempDir,
        job.Quality, 
        job.UserID, 
        job.VideoID.String(),
    )
    
    result.ProcessingTime = time.Since(startTime)
    
    if err != nil {
        result.Error = err.Error()
        return result
    }
    
    result.Success = true
    result.Quality = processedQuality
    result.SegmentCount = processedQuality.SegmentCount
    result.OutputSize = processedQuality.Size
    
    return result
}