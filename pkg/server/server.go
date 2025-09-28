package server

import (
	"net/http"
	"path/filepath"
)

// Server handles HTTP requests for video streaming
type Server struct {
	videoDir string
}

// New creates a new server instance
func New(videoDir string) *Server {
	return &Server{
		videoDir: videoDir,
	}
}

// ServeHTTP implements the http.Handler interface
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for streaming
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == "OPTIONS" {
		return
	}
	
	// Serve files from the video directory
	filePath := filepath.Join(s.videoDir, r.URL.Path)
	http.ServeFile(w, r, filePath)
}
