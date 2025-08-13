package database

import (
	"fmt"
	"log"
	"people-counting/internal/domain/entity"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Config holds database configuration
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewConnection establishes a connection to the PostgreSQL database
func NewConnection(config *Config) (*gorm.DB, error) {
	// Configure GORM logger
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level - Silent to disable query logs
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error
			Colorful:                  true,          // Color
		},
	)

	// Open connection to database
	db, err := gorm.Open(postgres.Open(config.DSN), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, err
	}

	// Get the underlying SQL DB object to set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool settings
	if config.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	} else {
		sqlDB.SetMaxOpenConns(25) // Default maximum open connections
	}

	if config.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	} else {
		sqlDB.SetMaxIdleConns(10) // Default maximum idle connections
	}

	if config.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		sqlDB.SetConnMaxLifetime(5 * time.Minute) // Default connection max lifetime
	}

	return db, nil
}

// IsTimescaleDBInstalled checks if TimescaleDB extension is installed
func IsTimescaleDBInstalled(db *gorm.DB) bool {
	var result int
	err := db.Raw("SELECT COUNT(*) FROM pg_extension WHERE extname = 'timescaledb'").Scan(&result).Error
	if err != nil || result == 0 {
		return false
	}
	return true
}

// MigrateDatabase handles database migrations using GORM
func MigrateDatabase(db *gorm.DB) error {
	// Check if TimescaleDB extension is installed
	if !IsTimescaleDBInstalled(db) {
		return fmt.Errorf("TimescaleDB extension is not installed")
	}

	// Check if tables already exist
	if !tablesExist(db) {
		log.Println("Creating database schema...")

		// Enable TimescaleDB extension
		if err := db.Exec("CREATE EXTENSION IF NOT EXISTS timescaledb").Error; err != nil {
			return fmt.Errorf("failed to enable TimescaleDB extension: %w", err)
		}

		// Auto migrate models
		if err := migrateModels(db); err != nil {
			return fmt.Errorf("failed to migrate models: %w", err)
		}

		// Create TimescaleDB hypertables
		if err := createHypertables(db); err != nil {
			return fmt.Errorf("failed to create hypertables: %w", err)
		}

		// Create indexes
		if err := createIndexes(db); err != nil {
			return fmt.Errorf("failed to create indexes: %w", err)
		}

		// Setup TimescaleDB features (compression, continuous aggregates)
		if err := setupTimescaleDBFeatures(db); err != nil {
			return fmt.Errorf("failed to setup TimescaleDB features: %w", err)
		}

		// Seed initial data
		if err := seedInitialData(db); err != nil {
			return fmt.Errorf("failed to seed initial data: %w", err)
		}

		log.Println("Database migration completed successfully")
	} else {
		log.Println("Tables already exist, skipping migration")
	}

	return nil
}

// tablesExist checks if the required tables already exist in the database
func tablesExist(db *gorm.DB) bool {
	var count int64
	tables := []string{"cameras", "people_counts", "alert_types", "alerts"}

	for _, table := range tables {
		result := db.Raw(`
			SELECT COUNT(*)
			FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = ?`, table).Scan(&count)

		if result.Error != nil || count == 0 {
			return false
		}
	}

	return true
}

// migrateModels uses GORM AutoMigrate to create the database schema
func migrateModels(db *gorm.DB) error {
	// Auto migrate models
	return db.AutoMigrate(
		&entity.Camera{},
		&entity.PeopleCount{},
		&entity.AlertType{},
		&entity.Alert{},
	)
}

// createHypertables creates TimescaleDB hypertables for time-series tables
func createHypertables(db *gorm.DB) error {
	// Create hypertable for people_counts
	if err := db.Exec("SELECT create_hypertable('people_counts', 'timestamp', if_not_exists => TRUE)").Error; err != nil {
		return err
	}

	return nil
}

// createIndexes creates indexes for better query performance
func createIndexes(db *gorm.DB) error {
	// Add indexes as defined in the SQL migration
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_people_counts_camera_id ON people_counts(camera_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_camera_id ON alerts(camera_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_alert_type_id ON alerts(alert_type_id)").Error; err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_alerts_is_active ON alerts(is_active)").Error; err != nil {
		return err
	}

	return nil
}

// setupTimescaleDBFeatures sets up advanced TimescaleDB features
func setupTimescaleDBFeatures(db *gorm.DB) error {
	// Activate compression for people_counts
	if err := db.Exec("ALTER TABLE people_counts SET (timescaledb.compress, timescaledb.compress_orderby = 'timestamp DESC')").Error; err != nil {
		return err
	}

	// Add compression policy
	if err := db.Exec("SELECT add_compression_policy('people_counts', INTERVAL '7 days')").Error; err != nil {
		return err
	}

	// Create continuous aggregate for hourly data
	if err := db.Exec(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS people_counts_hourly WITH (timescaledb.continuous) AS
		SELECT
			time_bucket('1 hour', timestamp) AS hour,
			camera_id,
			SUM(male_count) AS male_count,
			SUM(female_count) AS female_count,
			SUM(child_count) AS child_count,
			SUM(adult_count) AS adult_count,
			SUM(elderly_count) AS elderly_count,
			SUM(total_count) AS total_count
		FROM people_counts
		GROUP BY hour, camera_id
	`).Error; err != nil {
		return err
	}

	// Add policy for hourly continuous aggregate
	if err := db.Exec(`
		SELECT add_continuous_aggregate_policy('people_counts_hourly',
			start_offset => INTERVAL '2 days',
			end_offset => INTERVAL '1 hour',
			schedule_interval => INTERVAL '1 hour')
	`).Error; err != nil {
		return err
	}

	// Create continuous aggregate for daily data
	if err := db.Exec(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS people_counts_daily WITH (timescaledb.continuous) AS
		SELECT
			time_bucket('1 day', timestamp) AS day,
			camera_id,
			SUM(male_count) AS male_count,
			SUM(female_count) AS female_count,
			SUM(child_count) AS child_count,
			SUM(adult_count) AS adult_count,
			SUM(elderly_count) AS elderly_count,
			SUM(total_count) AS total_count
		FROM people_counts
		GROUP BY day, camera_id
	`).Error; err != nil {
		return err
	}

	// Add policy for daily continuous aggregate
	if err := db.Exec(`
		SELECT add_continuous_aggregate_policy('people_counts_daily',
			start_offset => INTERVAL '30 days',
			end_offset => INTERVAL '1 day',
			schedule_interval => INTERVAL '1 day')
	`).Error; err != nil {
		return err
	}

	return nil
}

// seedInitialData seeds initial data into the database
func seedInitialData(db *gorm.DB) error {
	// Seed alert types as defined in the 002_seed_alert_types.sql
	alertTypes := []entity.AlertType{
		{Name: "restricted", Icon: "restricted-area", Color: "#FF4D4F", Description: "Person detected in restricted area"},
		{Name: "smoke", Icon: "smoke", Color: "#FAAD14", Description: "Smoke detection alert"},
		{Name: "fire", Icon: "fire", Color: "#FF4D4F", Description: "Fire detection alert"},
	}

	for _, alertType := range alertTypes {
		if err := db.FirstOrCreate(&alertType, entity.AlertType{Name: alertType.Name}).Error; err != nil {
			return err
		}
	}

	return nil
}
