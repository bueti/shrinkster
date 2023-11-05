package model

import (
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Create a UrlModel struct which wraps the connection pool.
type UrlModel struct {
	DB *gorm.DB
}

type Url struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key" json:"id,omitempty"`
	Original string    `gorm:"type:varchar(2048);not null;uniqueIndex" json:"original"`
	ShortUrl string    `gorm:"type:varchar(11);not null;uniqueIndex" json:"short_url"`
	UserID   uuid.UUID `json:"user_id"`
	User     User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type UrlCreateRequest struct {
	Original  string    `json:"original" validate:"required,url"`
	ShortCode string    `json:"short_code" validate:"omitempty,alphanum,min=3,max=11"`
	UserId    uuid.UUID `json:"user_id"`
}

type UrlResponse struct {
	ID      uuid.UUID `json:"id"`
	FullUrl string    `json:"full_url"`
}

func (u *UrlModel) Create(urlReq *UrlCreateRequest) (Url, error) {
	url := new(Url)

	if urlReq.UserId == uuid.Nil {
		return Url{}, fmt.Errorf("user id is required")
	}

	if urlReq.ShortCode != "" {
		url.ShortUrl = urlReq.ShortCode
	} else {
		id := base62Encode(rand.Uint64())
		url.ShortUrl = id
	}

	url.Original = urlReq.Original
	url.UserID = urlReq.UserId

	result := u.DB.Create(url)
	if result.Error != nil {
		return Url{}, result.Error
	}

	return *url, nil
}

func (u *UrlModel) Get(shortUrl string) (Url, error) {
	url := new(Url)
	result := u.DB.Where("short_url = ?", shortUrl).First(&url)
	if result.Error != nil {
		return Url{}, result.Error
	}
	return *url, nil
}
