package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Initialize initializes the database connection with connection pooling (Golden Rule #9: Connection Pooling)
func Initialize(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is empty")
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for optimal performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)                  // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)                 // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Maximum lifetime of a connection
	sqlDB.SetConnMaxIdleTime(time.Minute * 10) // Maximum idle time of a connection

	return db, nil
}
