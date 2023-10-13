package models

import (
	"time"

	"github.com/google/uuid"
)

type ShortLink struct {
	ID          uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	SlashCode   string    `gorm:"not null;type:varchar(12);uniqueIndex;" json:"slash_code"`
	Origin      string    `gorm:"-:all" json:"origin"`
	Destination string    `gorm:"not null;type:varchar(512)" json:"destination"`
	Visitors    int       `json:"visitors"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
