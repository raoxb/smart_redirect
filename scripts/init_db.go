package main

import (
	"flag"
	"log"
	
	"golang.org/x/crypto/bcrypt"
	
	"github.com/raoxb/smart_redirect/internal/config"
	"github.com/raoxb/smart_redirect/internal/database"
	"github.com/raoxb/smart_redirect/internal/models"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/local.yaml", "Path to configuration file")
	flag.Parse()
	
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	
	db, err := database.NewPostgresDB(&cfg.Database.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	admin := &models.User{
		Username: "admin",
		Email:    "admin@smartredirect.com",
		Password: string(hashedPassword),
		Role:     "admin",
		IsActive: true,
	}
	
	var existingUser models.User
	if err := db.Where("username = ?", "admin").First(&existingUser).Error; err == nil {
		log.Println("Admin user already exists")
		return
	}
	
	if err := db.Create(admin).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}
	
	log.Println("Admin user created successfully")
	log.Println("Username: admin")
	log.Println("Password: admin123")
}