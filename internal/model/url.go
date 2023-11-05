package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Url struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key" json:"id,omitempty"`
	Original  string    `gorm:"type:varchar(2048);not null;uniqueIndex" json:"original"`
	ShortUrl  string    `gorm:"type:varchar(11);not null;uniqueIndex" json:"short_url"`
	ShortCode string    `json:"short_code,omitempty"`
	UserID    uuid.UUID `json:"user_id"`
	User      User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
