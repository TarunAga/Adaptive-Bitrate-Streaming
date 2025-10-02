package database

import (
    "fmt"
    "log"
    "os"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "github.com/TarunAga/adaptive-bitrate-streaming/pkg/entities"
)

// Config holds database configuration
type Config struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

// DB holds the database connection
var DB *gorm.DB

// GetDefaultConfig returns default PostgreSQL configuration
func GetDefaultConfig() *Config {
    return &Config{
        Host:     getEnv("DB_HOST", "localhost"),
        Port:     getEnv("DB_PORT", "5432"),
        User:     getEnv("DB_USER", "postgres"),
        Password: getEnv("DB_PASSWORD", "password"),
        DBName:   getEnv("DB_NAME", "adaptive_streaming"),
        SSLMode:  getEnv("DB_SSLMODE", "disable"),
    }
}

// Connect establishes database connection
func Connect(config *Config) error {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
        config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NowFunc: func() time.Time {
            return time.Now().UTC()
        },
        DisableForeignKeyConstraintWhenMigrating: true, // This helps with migration order
    })

    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }

    // Configure connection pool
    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("failed to get database instance: %w", err)
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    log.Println("Database connection established successfully")
    return nil
}

// AutoMigrate runs database migrations step by step
func AutoMigrate() error {
    if DB == nil {
        return fmt.Errorf("database not connected")
    }

    log.Println("Starting database migrations...")

    // Step 1: Drop existing tables if they exist (to start fresh)
    log.Println("Checking for existing tables...")
    if DB.Migrator().HasTable(&entities.Video{}) {
        log.Println("Dropping existing videos table...")
        if err := DB.Migrator().DropTable(&entities.Video{}); err != nil {
            log.Printf("Warning: Could not drop videos table: %v", err)
        }
    }
    
    if DB.Migrator().HasTable(&entities.User{}) {
        log.Println("Dropping existing users table...")
        if err := DB.Migrator().DropTable(&entities.User{}); err != nil {
            log.Printf("Warning: Could not drop users table: %v", err)
        }
    }

    // Step 2: Create Users table first (no foreign keys)
    log.Println("Creating users table...")
    err := DB.Migrator().CreateTable(&entities.User{})
    if err != nil {
        return fmt.Errorf("failed to create users table: %w", err)
    }
    log.Println("✅ Users table created successfully")

    // Step 3: Create Videos table (this will now work because users table exists)
    log.Println("Creating videos table...")
    err = DB.Migrator().CreateTable(&entities.Video{})
    if err != nil {
        return fmt.Errorf("failed to create videos table: %w", err)
    }
    log.Println("✅ Videos table created successfully")

    // Step 4: Create foreign key constraints manually if needed
    log.Println("Creating foreign key constraints...")
    err = createForeignKeys()
    if err != nil {
        log.Printf("Warning: Could not create foreign keys: %v", err)
        // Don't fail the migration for this
    }

    // Step 5: Create indexes for better performance
    log.Println("Creating database indexes...")
    err = createIndexes()
    if err != nil {
        log.Printf("Warning: Failed to create some indexes: %v", err)
    }

    log.Println("✅ Database migrations completed successfully")
    return nil
}

// createForeignKeys creates foreign key constraints manually
func createForeignKeys() error {
    // Check if foreign key already exists
    var count int64
    DB.Raw(`
        SELECT COUNT(*) 
        FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_videos_user_id' 
        AND table_name = 'videos'
    `).Scan(&count)

    if count == 0 {
        // Create foreign key constraint
        err := DB.Exec(`
            ALTER TABLE videos 
            ADD CONSTRAINT fk_videos_user_id 
            FOREIGN KEY (user_id) REFERENCES users(user_id) 
            ON DELETE CASCADE
        `).Error
        
        if err != nil {
            return fmt.Errorf("failed to create foreign key constraint: %w", err)
        }
        log.Println("✅ Foreign key constraint created")
    } else {
        log.Println("✅ Foreign key constraint already exists")
    }

    return nil
}

// createIndexes creates additional indexes for better performance
func createIndexes() error {
    indexes := []struct {
        name  string
        query string
    }{
        {
            name:  "idx_videos_video_id",
            query: "CREATE INDEX IF NOT EXISTS idx_videos_video_id ON videos(video_id)",
        },
        {
            name:  "idx_videos_user_id", 
            query: "CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos(user_id)",
        },
        {
            name:  "idx_users_username",
            query: "CREATE INDEX IF NOT EXISTS idx_users_username ON users(user_name)",
        },
        {
            name:  "idx_videos_status",
            query: "CREATE INDEX IF NOT EXISTS idx_videos_status ON videos(status)",
        },
    }

    for _, idx := range indexes {
        err := DB.Exec(idx.query).Error
        if err != nil {
            log.Printf("Warning: Failed to create index %s: %v", idx.name, err)
        } else {
            log.Printf("✅ Index %s created", idx.name)
        }
    }

    return nil
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
    return DB
}

// Close closes the database connection
func Close() error {
    if DB != nil {
        sqlDB, err := DB.DB()
        if err != nil {
            return err
        }
        return sqlDB.Close()
    }
    return nil
}

// ResetDatabase drops and recreates all tables (useful for development)
func ResetDatabase() error {
    if DB == nil {
        return fmt.Errorf("database not connected")
    }

    log.Println("🔄 Resetting database...")
    return AutoMigrate()
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}