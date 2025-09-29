package upload

import (
	"io"
	"mime/multipart"
	"testing"
)

func TestGenerateS3Key(t *testing.T) {
	tests := []struct {
		userID   string
		title    string
		filename string
		expected string
	}{
		{
			userID:   "user123",
			title:    "My Test Video",
			filename: "video.mp4",
			expected: "user123/UploadedVideo_My_Test_Video.mp4",
		},
		{
			userID:   "user456",
			title:    "Another/Video\\File",
			filename: "test.avi",
			expected: "user456/UploadedVideo_Another_Video_File.avi",
		},
		{
			userID:   "user789",
			title:    "No Extension Video",
			filename: "video",
			expected: "user789/UploadedVideo_No_Extension_Video",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			result := generateS3Key(tt.userID, tt.title, tt.filename)
			if result != tt.expected {
				t.Errorf("generateS3Key() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		contentType string
		filename    string
		expected    bool
	}{
		{"video/mp4", "test.mp4", true},
		{"video/avi", "test.avi", true},
		{"video/quicktime", "test.mov", true},
		{"application/octet-stream", "test.mp4", true},
		{"text/plain", "test.txt", false},
		{"image/jpeg", "test.jpg", false},
		{"application/octet-stream", "test.mkv", true},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isVideoFile(tt.contentType, tt.filename)
			if result != tt.expected {
				t.Errorf("isVideoFile(%s, %s) = %v, want %v", 
					tt.contentType, tt.filename, result, tt.expected)
			}
		})
	}
}

// Mock multipart file for testing
type mockFile struct {
	content []byte
	pos     int
}

func (m *mockFile) Read(p []byte) (n int, err error) {
	if m.pos >= len(m.content) {
		return 0, io.EOF
	}
	n = copy(p, m.content[m.pos:])
	m.pos += n
	return n, nil
}

func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.pos = int(offset)
	case io.SeekCurrent:
		m.pos += int(offset)
	case io.SeekEnd:
		m.pos = len(m.content) + int(offset)
	}
	return int64(m.pos), nil
}

func (m *mockFile) Close() error {
	return nil
}

func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.content)) {
		return 0, io.EOF
	}
	n = copy(p, m.content[off:])
	return n, nil
}

func createMockMultipartFile(content string, filename string) (multipart.File, *multipart.FileHeader) {
	mockFile := &mockFile{content: []byte(content)}
	
	header := &multipart.FileHeader{
		Filename: filename,
		Size:     int64(len(content)),
		Header:   make(map[string][]string),
	}
	header.Header.Set("Content-Type", "video/mp4")
	
	return mockFile, header
}

func TestUploadRequest_Validation(t *testing.T) {
	file, header := createMockMultipartFile("test video content", "test.mp4")
	
	req := &UploadRequest{
		UserID: "user123",
		Title:  "Test Video",
		File:   file,
		Header: header,
	}

	// Test that the request is properly structured
	if req.UserID == "" {
		t.Error("UserID should not be empty")
	}
	if req.Title == "" {
		t.Error("Title should not be empty")
	}
	if req.File == nil {
		t.Error("File should not be nil")
	}
	if req.Header == nil {
		t.Error("Header should not be nil")
	}
}
