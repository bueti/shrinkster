package model

import "gorm.io/gorm"

type Url struct {
	gorm.Model
	ShortURL  string `json:"short_url"`
	Original  string `json:"original"`
	Visits    uint   `json:"visits"`
	ShortCode string `json:"short_code"`
	UserID    int    `json:"user_id"`
	User      User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
