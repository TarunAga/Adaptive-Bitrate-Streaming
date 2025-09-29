# Test script for video upload API (PowerShell)
# Usage: .\test-upload.ps1 [video-file-path]

param(
    [string]$VideoFile = "sample-video.mp4"
)

$ApiUrl = "http://localhost:8081/api/v1"
$UserId = "user123"
$Title = "Test Video Upload"

Write-Host "Testing Video Upload API" -ForegroundColor Green
Write-Host "=========================" -ForegroundColor Green

# Check if video file exists
if (-not (Test-Path $VideoFile)) {
    Write-Host "Error: Video file '$VideoFile' not found!" -ForegroundColor Red
    Write-Host "Usage: .\test-upload.ps1 [path-to-video-file]"
    exit 1
}

Write-Host "Video file: $VideoFile"
Write-Host "User ID: $UserId"
Write-Host "Title: $Title"
Write-Host ""

# Test health check
Write-Host "1. Testing health check..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$ApiUrl/health" -Method Get
    $healthResponse | ConvertTo-Json
} catch {
    Write-Host "Health check failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test upload info
Write-Host "2. Getting upload info..." -ForegroundColor Yellow
try {
    $infoResponse = Invoke-RestMethod -Uri "$ApiUrl/upload/info" -Method Get
    $infoResponse | ConvertTo-Json
} catch {
    Write-Host "Upload info failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test video upload
Write-Host "3. Uploading video..." -ForegroundColor Yellow
try {
    # Create multipart form data
    $boundary = [System.Guid]::NewGuid().ToString()
    $LF = "`r`n"
    
    $fileContent = [System.IO.File]::ReadAllBytes($VideoFile)
    $fileName = [System.IO.Path]::GetFileName($VideoFile)
    
    $bodyLines = (
        "--$boundary",
        "Content-Disposition: form-data; name=`"userId`"$LF",
        $UserId,
        "--$boundary",
        "Content-Disposition: form-data; name=`"title`"$LF",
        $Title,
        "--$boundary",
        "Content-Disposition: form-data; name=`"video`"; filename=`"$fileName`"",
        "Content-Type: video/mp4$LF"
    ) -join $LF
    
    $bodyLines += $LF
    
    # Convert to bytes
    $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($bodyLines)
    $endBytes = [System.Text.Encoding]::UTF8.GetBytes("$LF--$boundary--$LF")
    
    # Combine all bytes
    $fullBody = New-Object byte[] ($bodyBytes.Length + $fileContent.Length + $endBytes.Length)
    [Array]::Copy($bodyBytes, 0, $fullBody, 0, $bodyBytes.Length)
    [Array]::Copy($fileContent, 0, $fullBody, $bodyBytes.Length, $fileContent.Length)
    [Array]::Copy($endBytes, 0, $fullBody, $bodyBytes.Length + $fileContent.Length, $endBytes.Length)
    
    $headers = @{
        'Content-Type' = "multipart/form-data; boundary=$boundary"
    }
    
    $uploadResponse = Invoke-RestMethod -Uri "$ApiUrl/upload" -Method Post -Body $fullBody -Headers $headers
    $uploadResponse | ConvertTo-Json
    
} catch {
    Write-Host "Upload failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Upload test completed!" -ForegroundColor Green
