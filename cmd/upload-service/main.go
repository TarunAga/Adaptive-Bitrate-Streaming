package main

import (
	"log"
	"net/http"
	"os"

	"github.com/TarunAga/adaptive-bitrate-streaming/pkg/upload"
	"github.com/gorilla/mux"
)

func main() {
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Initialize upload service
	uploadService, err := upload.NewService()
	if err != nil {
		log.Fatalf("Failed to initialize upload service: %v", err)
	}

	// Initialize handler
	uploadHandler := upload.NewHandler(uploadService)

	// Setup routes
	router := mux.NewRouter()
	
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/upload", uploadHandler.UploadVideoHandler).Methods("POST", "OPTIONS")
	api.HandleFunc("/upload/info", uploadHandler.GetUploadInfoHandler).Methods("GET")
	api.HandleFunc("/health", uploadHandler.HealthCheckHandler).Methods("GET")

	// Add logging middleware
	router.Use(loggingMiddleware)

	log.Printf("Starting upload service on port %s", port)
	log.Printf("Upload endpoint: http://localhost:%s/api/v1/upload", port)
	log.Printf("Health check: http://localhost:%s/api/v1/health", port)
	log.Printf("Upload info: http://localhost:%s/api/v1/upload/info", port)
	
	// Start server
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
