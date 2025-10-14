package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize initializes the database connection with connection pooling (Golden Rule #9: Connection Pooling)
func Initialize(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is empty - check POSTGRES_* environment variables")
	}

	log.Println("  📡 Connecting to PostgreSQL database...")

	// Configure GORM with custom logger for better observability
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // Log warnings and errors only
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("  ✅ Database connection established")

	// Configure connection pool for optimal performance
	log.Println("  ⚙️  Configuring connection pool...")
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)                  // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(time.Minute * 10) // Maximum idle time of a connection

	log.Println("  ✅ Connection pool configured (MaxIdle: 10, MaxOpen: 100)")

	// Test connection
	log.Println("  🔍 Testing database connectivity...")
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	log.Println("  ✅ Database ping successful")

	return db, nil
}
