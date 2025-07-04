package models

import (
	"time"
	"gorm.io/gorm"
)

type Link struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	LinkID      string    `gorm:"uniqueIndex;size:10" json:"link_id"`
	BusinessUnit string   `gorm:"size:10" json:"business_unit"`
	Network     string    `gorm:"size:50" json:"network"`
	TotalCap    int       `json:"total_cap"`
	CurrentHits int       `json:"current_hits"`
	BackupURL   string    `json:"backup_url"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	
	Targets []Target `gorm:"foreignKey:LinkID;references:ID" json:"targets,omitempty"`
	Permissions []LinkPermission `gorm:"foreignKey:LinkID;references:ID" json:"permissions,omitempty"`
}

type Target struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	LinkID        uint      `gorm:"index" json:"link_id"`
	URL           string    `json:"url"`
	Weight        int       `json:"weight"`
	Cap           int       `json:"cap"`
	CurrentHits   int       `json:"current_hits"`
	Countries     string    `json:"countries"`
	ParamMapping  string    `gorm:"type:jsonb" json:"param_mapping"`
	StaticParams  string    `gorm:"type:jsonb" json:"static_params"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	
	Link *Link `gorm:"foreignKey:LinkID" json:"link,omitempty"`
}