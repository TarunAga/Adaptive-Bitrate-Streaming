# AWS Configuration for Upload Service

## Prerequisites

1. **AWS Account**: You need an active AWS account
2. **AWS Credentials**: Configure your AWS credentials using one of these methods:

### Method 1: AWS CLI Configuration
```bash
aws configure
```
This will prompt for:
- AWS Access Key ID
- AWS Secret Access Key
- Default region name (e.g., us-east-1)
- Default output format (json)

### Method 2: Environment Variables
```bash
export AWS_ACCESS_KEY_ID=your_access_key_here
export AWS_SECRET_ACCESS_KEY=your_secret_key_here
export AWS_DEFAULT_REGION=us-east-1
```

### Method 3: AWS Profile in ~/.aws/credentials
```ini
[default]
aws_access_key_id = your_access_key_here
aws_secret_access_key = your_secret_key_here
region = us-east-1
```

## S3 Bucket

The service uses the bucket name: `adaptive-bitrate-streaming-videos`

### Create S3 Bucket (Optional)
You can create the bucket manually in AWS Console or programmatically:

```bash
aws s3 mb s3://adaptive-bitrate-streaming-videos --region us-east-1
```

### Bucket Policy for Public Read (Optional)
If you want uploaded videos to be publicly accessible:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicReadGetObject",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::adaptive-bitrate-streaming-videos/*"
        }
    ]
}
```

## IAM Permissions

Your AWS user/role needs the following S3 permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:PutObjectAcl",
                "s3:GetObject",
                "s3:DeleteObject"
            ],
            "Resource": "arn:aws:s3:::adaptive-bitrate-streaming-videos/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::adaptive-bitrate-streaming-videos"
        }
    ]
}
```

## Configuration Variables

You can modify these constants in `pkg/upload/service.go`:

- `BucketName`: S3 bucket name
- `AWSRegion`: AWS region
- File size limit (currently 500MB)
- Allowed video formats

## Testing

Use curl to test the upload:

```bash
curl -X POST \
  -F "userId=user123" \
  -F "title=My Test Video" \
  -F "video=@/path/to/your/video.mp4" \
  http://localhost:8081/api/v1/upload
```
