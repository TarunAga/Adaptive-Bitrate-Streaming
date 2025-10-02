package main

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/database"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/upload"
    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/auth"
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

    // Create handlers
    uploadHandler := upload.NewHandler(uploadService)
    authHandler := auth.NewAuthHandler(database.GetDB())

    // Setup routes
    router := mux.NewRouter()
    
    // Auth routes (no authentication required)
    router.HandleFunc("/api/v1/auth/register", authHandler.RegisterHandler).Methods("POST", "OPTIONS")
    router.HandleFunc("/api/v1/auth/login", authHandler.LoginHandler).Methods("POST", "OPTIONS")
    
    // Protected upload routes (authentication required)
    router.HandleFunc("/api/v1/upload", authHandler.AuthMiddleware(uploadHandler.UploadVideoHandler)).Methods("POST", "OPTIONS")

    // Public info routes
    router.HandleFunc("/api/v1/upload/info", uploadHandler.GetUploadInfoHandler).Methods("GET")
    router.HandleFunc("/api/v1/health", uploadHandler.HealthCheckHandler).Methods("GET")

    // ✅ FIXED: Serve static files (choose ONE folder)
    router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
    
    // ✅ Root redirect to our test page
    router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/static/index.html", http.StatusMovedPermanently)
    }).Methods("GET")

    log.Printf("Upload service starting on :8081")
    log.Printf("Database: PostgreSQL connected successfully")
    log.Printf("Routes registered:")
    log.Printf("  POST /api/v1/auth/register")
    log.Printf("  POST /api/v1/auth/login")
    log.Printf("  POST /api/v1/upload (protected)")
    log.Printf("  GET  /api/v1/upload/info")
    log.Printf("  GET  /api/v1/health")
    log.Printf("  GET  /static/ (static files)")
    log.Printf("  GET  / (redirects to /static/index.html)")
    log.Printf("Server listening on http://localhost:8081")

    log.Fatal(http.ListenAndServe(":8081", router))
}