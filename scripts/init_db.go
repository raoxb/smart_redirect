package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	
	"github.com/raoxb/smart_redirect/internal/config"
	"golang.org/x/crypto/bcrypt"
)

// 临时模型，避免外键问题
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50" json:"username"`
	Email     string    `gorm:"uniqueIndex;size:100" json:"email"`
	Password  string    `json:"-"`
	Role      string    `gorm:"size:20;default:user" json:"role"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Link struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	LinkID       string    `gorm:"uniqueIndex;size:10" json:"link_id"`
	BusinessUnit string    `gorm:"size:10" json:"business_unit"`
	Network      string    `gorm:"size:50" json:"network"`
	TotalCap     int       `json:"total_cap"`
	CurrentHits  int       `json:"current_hits"`
	BackupURL    string    `json:"backup_url"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Target struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	LinkID       uint      `gorm:"index" json:"link_id"`
	URL          string    `json:"url"`
	Weight       int       `json:"weight"`
	Cap          int       `json:"cap"`
	CurrentHits  int       `json:"current_hits"`
	Countries    string    `json:"countries"`
	ParamMapping string    `gorm:"type:jsonb" json:"param_mapping"`
	StaticParams string    `gorm:"type:jsonb" json:"static_params"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AccessLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	LinkID    uint      `gorm:"index" json:"link_id"`
	TargetID  uint      `gorm:"index" json:"target_id"`
	IP        string    `gorm:"size:45" json:"ip"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	Referer   string    `gorm:"size:500" json:"referer"`
	Country   string    `gorm:"size:2" json:"country"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	// Load config
	cfg, err := config.Load("config/local.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to database directly without auto migration
	db, err := gorm.Open(postgres.Open(cfg.Database.Postgres.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Drop all tables (for clean start)
	log.Println("Dropping existing tables...")
	db.Exec("DROP TABLE IF EXISTS link_permissions CASCADE")
	db.Exec("DROP TABLE IF EXISTS access_logs CASCADE")
	db.Exec("DROP TABLE IF EXISTS targets CASCADE") 
	db.Exec("DROP TABLE IF EXISTS links CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")
	db.Exec("DROP TABLE IF EXISTS link_templates CASCADE")

	// Create tables in correct order
	log.Println("Creating tables...")
	
	err = db.AutoMigrate(
		&User{},
		&Link{},
		&Target{},
		&AccessLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create default admin user
	log.Println("Creating default admin user...")
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	
	admin := User{
		Username: "admin",
		Email:    "admin@smartredirect.com",
		Password: string(hashedPassword),
		Role:     "admin",
		IsActive: true,
	}
	
	db.Create(&admin)

	// Create sample data
	log.Println("Creating sample data...")
	
	// Sample links
	link1 := Link{
		LinkID:       "abc123",
		BusinessUnit: "bu01",
		Network:      "mi",
		TotalCap:     1000,
		BackupURL:    "https://backup.example.com",
		IsActive:     true,
	}
	db.Create(&link1)

	link2 := Link{
		LinkID:       "def456", 
		BusinessUnit: "bu02",
		Network:      "google",
		TotalCap:     2000,
		BackupURL:    "https://backup2.example.com",
		IsActive:     true,
	}
	db.Create(&link2)

	// Sample targets
	targets := []Target{
		{
			LinkID:       link1.ID,
			URL:          "https://target1.example.com",
			Weight:       70,
			Cap:          500,
			Countries:    `["US","CA"]`,
			ParamMapping: `{"kw":"q"}`,
			StaticParams: `{"ref":"test"}`,
			IsActive:     true,
		},
		{
			LinkID:       link1.ID,
			URL:          "https://target2.example.com",
			Weight:       30,
			Cap:          300,
			Countries:    `["UK","DE"]`,
			IsActive:     true,
		},
		{
			LinkID:       link2.ID,
			URL:          "https://target3.example.com",
			Weight:       50,
			Cap:          800,
			Countries:    `["US"]`,
			ParamMapping: `{"keyword":"search"}`,
			StaticParams: `{"source":"redirect"}`,
			IsActive:     true,
		},
		{
			LinkID:       link2.ID,
			URL:          "https://target4.example.com",
			Weight:       50,
			Cap:          800,
			Countries:    `["ALL"]`,
			IsActive:     true,
		},
	}

	for _, target := range targets {
		db.Create(&target)
	}

	log.Println("Database initialization completed successfully!")
}