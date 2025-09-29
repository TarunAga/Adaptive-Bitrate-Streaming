#!/bin/bash

# Test script for video upload API
# Usage: ./test-upload.sh [video-file-path]

API_URL="http://localhost:8081/api/v1"
VIDEO_FILE=${1:-"sample-video.mp4"}
USER_ID="user123"
TITLE="Test Video Upload"

echo "Testing Video Upload API"
echo "========================="

# Check if video file exists
if [ ! -f "$VIDEO_FILE" ]; then
    echo "Error: Video file '$VIDEO_FILE' not found!"
    echo "Usage: $0 [path-to-video-file]"
    exit 1
fi

echo "Video file: $VIDEO_FILE"
echo "User ID: $USER_ID"
echo "Title: $Title"
echo ""

# Test health check
echo "1. Testing health check..."
curl -s "$API_URL/health" | jq .
echo ""

# Test upload info
echo "2. Getting upload info..."
curl -s "$API_URL/upload/info" | jq .
echo ""

# Test video upload
echo "3. Uploading video..."
curl -X POST \
  -F "userId=$USER_ID" \
  -F "title=$TITLE" \
  -F "video=@$VIDEO_FILE" \
  "$API_URL/upload" | jq .

echo ""
echo "Upload test completed!"
