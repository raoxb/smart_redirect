package main

import (
	"fmt"
	"log"
	
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=password dbname=smart_redirect port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// Update admin user password
	result := db.Exec("UPDATE users SET password = ? WHERE username = ?", string(hashedPassword), "admin")
	if result.Error != nil {
		log.Fatal("Failed to update password:", result.Error)
	}

	fmt.Printf("Admin user password updated. Affected rows: %d\n", result.RowsAffected)
	fmt.Println("Username: admin")
	fmt.Println("Password: admin123")
}