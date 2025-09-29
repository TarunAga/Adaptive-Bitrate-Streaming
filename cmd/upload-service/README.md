# Video Upload Service

A Go-based REST API service for uploading video files to Amazon S3 with multipart form support.

## Features

- ✅ Multipart file upload support
- ✅ S3 integration with AWS SDK
- ✅ Video file validation (content type and extension)
- ✅ Configurable file size limits (default: 500MB)
- ✅ CORS support for web clients
- ✅ Health check endpoint
- ✅ Structured logging
- ✅ Custom S3 key generation pattern

## API Endpoints

### Upload Video
- **POST** `/api/v1/upload`
- **Content-Type**: `multipart/form-data`
- **Fields**:
  - `userId` (string, required): User identifier
  - `title` (string, required): Video title
  - `video` (file, required): Video file to upload

**Example Response**:
```json
{
  "success": true,
  "message": "Video uploaded successfully",
  "s3_key": "user123/UploadedVideo_My_Test_Video.mp4",
  "s3_url": "https://adaptive-bitrate-streaming-videos.s3.us-east-1.amazonaws.com/user123/UploadedVideo_My_Test_Video.mp4",
  "file_size": 1048576
}
```

### Get Upload Info
- **GET** `/api/v1/upload/info`

**Response**:
```json
{
  "max_file_size_mb": 500,
  "allowed_formats": ["mp4", "avi", "mov", "mkv", "webm"],
  "bucket_name": "adaptive-bitrate-streaming-videos",
  "required_fields": ["userId", "title", "video"]
}
```

### Health Check
- **GET** `/api/v1/health`

**Response**:
```json
{
  "status": "healthy",
  "service": "upload-service"
}
```

## S3 Key Pattern

Uploaded files are stored in S3 with the following key pattern:
```
{userId}/UploadedVideo_{title}.{extension}
```

**Examples**:
- `user123/UploadedVideo_My_Test_Video.mp4`
- `user456/UploadedVideo_Another_Video.avi`

Special characters in titles are replaced with underscores for filesystem safety.

## Setup and Installation

### Prerequisites

1. **Go 1.19+** installed
2. **AWS Account** with S3 access
3. **AWS Credentials** configured (see [AWS_SETUP.md](../AWS_SETUP.md))

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the service:
   ```bash
   go build -o bin/upload-service ./cmd/upload-service
   ```

### Configuration

Edit the constants in `pkg/upload/service.go`:

```go
const (
    BucketName = "your-bucket-name"     // Change to your S3 bucket
    AWSRegion  = "us-east-1"            // Change to your AWS region
)
```

### Running the Service

```bash
# Default port 8081
./bin/upload-service

# Custom port
PORT=8080 ./bin/upload-service
```

The service will start on `http://localhost:8081` by default.

## Testing

### Option 1: Web Interface
1. Start the service
2. Open `testdata/upload-test.html` in your browser
3. Fill in the form and upload a video file

### Option 2: cURL
```bash
curl -X POST \
  -F "userId=user123" \
  -F "title=My Test Video" \
  -F "video=@/path/to/your/video.mp4" \
  http://localhost:8081/api/v1/upload
```

### Option 3: PowerShell Script
```powershell
.\scripts\test-upload.ps1 "path\to\your\video.mp4"
```

### Option 4: Unit Tests
```bash
go test ./pkg/upload/...
```

## File Validation

The service validates uploaded files based on:

### Content Types
- `video/mp4`
- `video/avi`
- `video/quicktime`
- `video/x-msvideo`
- `video/x-matroska`
- `video/webm`

### File Extensions
- `.mp4`
- `.avi`
- `.mov`
- `.mkv`
- `.webm`

### Size Limits
- Maximum file size: 500MB (configurable)

## Error Handling

The API returns structured error responses:

```json
{
  "success": false,
  "message": "Error description"
}
```

Common error scenarios:
- Missing required fields (400)
- Invalid file type (400)
- File too large (400)
- S3 upload failure (500)

## AWS Permissions

Your AWS credentials need the following S3 permissions:
- `s3:PutObject`
- `s3:PutObjectAcl`
- `s3:GetObject`
- `s3:ListBucket`

See [AWS_SETUP.md](../AWS_SETUP.md) for detailed configuration.

## Monitoring and Logging

The service includes:
- Request logging middleware
- Structured error logging
- Health check endpoint for monitoring

## Development

### Project Structure
```
cmd/upload-service/
├── main.go              # Application entry point
pkg/upload/
├── service.go           # S3 upload service
├── handler.go           # HTTP handlers
└── service_test.go      # Unit tests
```

### Adding Features

To extend the service:
1. Add new endpoints in `pkg/upload/handler.go`
2. Implement business logic in `pkg/upload/service.go`
3. Add tests in `pkg/upload/service_test.go`
4. Update routes in `cmd/upload-service/main.go`

## Security Considerations

- Configure proper S3 bucket policies
- Use IAM roles with minimal required permissions
- Consider adding authentication/authorization
- Validate file content, not just extensions
- Set appropriate CORS policies
- Monitor upload patterns for abuse
