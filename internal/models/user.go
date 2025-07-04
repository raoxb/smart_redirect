package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50" json:"username"`
	Email     string    `gorm:"uniqueIndex;size:100" json:"email"`
	Password  string    `json:"-"`
	Role      string    `gorm:"size:20;default:'user'" json:"role"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	Permissions []LinkPermission `gorm:"foreignKey:UserID;references:ID" json:"permissions,omitempty"`
}

type LinkPermission struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"`
	LinkID    uint      `gorm:"index" json:"link_id"`
	CanEdit   bool      `gorm:"default:false" json:"can_edit"`
	CanDelete bool      `gorm:"default:false" json:"can_delete"`
	CreatedAt time.Time `json:"created_at"`
	
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Link *Link `gorm:"foreignKey:LinkID" json:"link,omitempty"`
}