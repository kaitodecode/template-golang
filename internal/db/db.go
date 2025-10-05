// pkg/db/db.go
package db

import (
	"fmt"
	"sync"
	"time"

	"template-golang/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var (
	db      *gorm.DB
	once    sync.Once
	initErr error
)

// ConnectDB establishes a singleton connection to PostgreSQL using GORM
func ConnectDB() (*gorm.DB, error) {
	once.Do(func() {
		cfg := config.GetConfig()

		// GORM configuration
		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, // Use singular table names
			},
			PrepareStmt: true, // Cache prepared statements for better performance
		}

		// Connect to database
		var err error
		db, err = gorm.Open(postgres.Open(cfg.DBURL), gormConfig)
		if err != nil {
			initErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// Configure connection pool
		sqlDB, err := db.DB()
		if err != nil {
			initErr = fmt.Errorf("failed to get database connection: %w", err)
			return
		}

		// Set connection pool settings
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(15 * time.Minute)
	})

	return db, initErr
}

// GetDB returns the GORM database instance
func GetDB() *gorm.DB {
	return db
}

// CloseDB closes the database connection safely
func CloseDB() {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}
