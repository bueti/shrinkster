package model

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	Token  string    `gorm:"size:43;uniqueIndex"`
	Data   []byte    `gorm:"not null"`
	Expiry time.Time `gorm:"not null;index"`
}
