package database

import (
	"fmt"
	"time"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/raoxb/smart_redirect/internal/api"
	"github.com/raoxb/smart_redirect/internal/config"
	"github.com/raoxb/smart_redirect/internal/models"
)

func NewPostgresDB(cfg *config.PostgresConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	
	// Skip auto migration - database is already initialized
	// if err := autoMigrate(db); err != nil {
	//	return nil, fmt.Errorf("failed to migrate database: %w", err)
	// }
	
	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Link{},
		&models.Target{},
		&models.LinkPermission{},
		&models.AccessLog{},
		&api.LinkTemplate{},
	)
}