package models

import (
	"time"
)

type AccessLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	LinkID     uint      `gorm:"index" json:"link_id"`
	TargetID   uint      `gorm:"index" json:"target_id"`
	IP         string    `gorm:"index;size:45" json:"ip"`
	UserAgent  string    `json:"user_agent"`
	Referer    string    `json:"referer"`
	Country    string    `gorm:"size:2" json:"country"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
	
	Link   *Link   `gorm:"foreignKey:LinkID;references:ID" json:"link,omitempty"`
	Target *Target `gorm:"foreignKey:TargetID;references:ID" json:"target,omitempty"`
}