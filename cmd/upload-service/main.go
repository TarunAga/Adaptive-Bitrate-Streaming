package main

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/rs/cors"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/database"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/upload"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/auth"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/streaming" // ✅ ADD: Import streaming package
    "github.com/joho/godotenv"
)

func main() {
    // Initialize database
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using system environment variables")
    }
    
    log.Println("Connecting to PostgreSQL database...")
    dbConfig := database.GetDefaultConfig()
    err = database.Connect(dbConfig)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer database.Close()

    // Run migrations
    log.Println("Running database migrations...")
    err = database.AutoMigrate()
    if err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }

    // Create services
    uploadService, err := upload.NewService(database.GetDB())
    if err != nil {
        log.Fatalf("Failed to create upload service: %v", err)
    }

    // ✅ ADD: Create streaming service
    streamingService, err := streaming.NewService(database.GetDB())
    if err != nil {
        log.Fatalf("Failed to create streaming service: %v", err)
    }

    // Create handlers
    uploadHandler := upload.NewHandler(uploadService)
    authHandler := auth.NewAuthHandler(database.GetDB())
    streamingHandler := streaming.NewHandler(streamingService) // ✅ ADD: Create streaming handler

    // Setup routes
    router := mux.NewRouter()
    
    // Auth routes (no authentication required)
    router.HandleFunc("/api/v1/auth/register", authHandler.RegisterHandler).Methods("POST", "OPTIONS")
    router.HandleFunc("/api/v1/auth/login", authHandler.LoginHandler).Methods("POST", "OPTIONS")
    
    // Protected upload routes (authentication required)
    router.HandleFunc("/api/v1/upload", authHandler.AuthMiddleware(uploadHandler.UploadVideoHandler)).Methods("POST", "OPTIONS")
    router.HandleFunc("/api/v1/videos", authHandler.AuthMiddleware(uploadHandler.GetUserVideosHandler)).Methods("GET", "OPTIONS")

    // ✅ ADD: Protected streaming routes (authentication required)
    router.HandleFunc("/api/v1/video/{videoId}/stream", authHandler.AuthMiddleware(streamingHandler.GetVideoStreamHandler)).Methods("GET", "OPTIONS")

    // Public info routes
    router.HandleFunc("/api/v1/upload/info", uploadHandler.GetUploadInfoHandler).Methods("GET")
    router.HandleFunc("/api/v1/health", uploadHandler.HealthCheckHandler).Methods("GET")

    // Setup basic CORS for API
    c := cors.New(cors.Options{
        AllowedOrigins: []string{"*"},
        AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{"*"},
        AllowCredentials: true,
    })

    handler := c.Handler(router)

    log.Printf("🚀 Adaptive Bitrate Streaming API starting...")
    log.Printf("📊 Database: PostgreSQL connected successfully")
    log.Printf("📡 API Routes:")
    log.Printf("  POST /api/v1/auth/register")
    log.Printf("  POST /api/v1/auth/login")
    log.Printf("  POST /api/v1/upload (protected)")
    log.Printf("  GET  /api/v1/videos (protected)")
    log.Printf("  GET  /api/v1/video/{videoId}/stream (protected)") // ✅ ADD: Log new streaming route
    log.Printf("  GET  /api/v1/upload/info")
    log.Printf("  GET  /api/v1/health")
    log.Printf("✅ API Server ready at http://localhost:8081")

    log.Fatal(http.ListenAndServe(":8081", handler))
}