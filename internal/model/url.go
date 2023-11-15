package model

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

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
	UserID   uuid.UUID `gorm:"type:uuid" json:"user_id"`
	User     User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Visits   int       `gorm:"default:0" json:"visits"`
}

type UrlCreateRequest struct {
	Original  string    `json:"original" validate:"required,url"`
	ShortCode string    `json:"short_code" validate:"omitempty,alphanum,min=3,max=11"`
	UserID    uuid.UUID `json:"user_id"`
}

type UrlResponse struct {
	ID      uuid.UUID `json:"id"`
	FullUrl string    `json:"full_url"`
}

type UrlByUserRequest struct {
	ID uuid.UUID `json:"user_id"`
}

type UrlByUserResponse struct {
	ID        uuid.UUID `json:"id"`
	Original  string    `json:"original"`
	ShortUrl  string    `json:"short_url"`
	Visits    int       `json:"visits"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (u *UrlModel) Create(urlReq *UrlCreateRequest) (Url, error) {
	url := new(Url)

	if urlReq.UserID == uuid.Nil {
		return Url{}, fmt.Errorf("user id is required")
	}

	if strings.Contains(urlReq.Original, "shrink.ch/s/") {
		return Url{}, fmt.Errorf("url cannot start with shrink.ch/s/")
	}
	if urlReq.ShortCode != "" {
		url.ShortUrl = urlReq.ShortCode
	} else {
		id := base62Encode(rand.Uint64())
		url.ShortUrl = id
	}

	url.Original = urlReq.Original
	url.UserID = urlReq.UserID

	result := u.DB.Create(url)
	if result.Error != nil {
		return Url{}, result.Error
	}

	return *url, nil
}

func (u *UrlModel) GetRedirect(shortUrl string) (Url, error) {
	url := new(Url)
	result := u.DB.Where("short_url = ?", shortUrl).First(&url)
	if result.Error != nil {
		return Url{}, result.Error
	}

	go func() {
		u.DB.Model(&url).Update("visits", gorm.Expr("visits + 1"))
	}()

	return *url, nil
}

// GetUrlByUser returns all URLs for a given user
func (u *UrlModel) GetUrlByUser(userId uuid.UUID) (*[]UrlByUserResponse, error) {
	var urls []Url
	result := u.DB.Where("user_id = ?", userId).Find(&urls)
	if result.Error != nil {
		return nil, result.Error
	}

	resp := []UrlByUserResponse{}
	for _, url := range urls {
		resp = append(resp, UrlByUserResponse{
			ID:        url.ID,
			Original:  url.Original,
			ShortUrl:  url.ShortUrl,
			Visits:    url.Visits,
			CreatedAt: url.CreatedAt,
			UpdatedAt: url.UpdatedAt,
		})
	}

	return &resp, nil
}

func (u *UrlModel) Delete(urlUUID uuid.UUID) error {
	url := new(Url)
	result := u.DB.Where("id = ?", urlUUID).Delete(&url)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (u *UrlModel) Find(urlUUID uuid.UUID) *Url {
	url := new(Url)
	result := u.DB.Where("id = ?", urlUUID).First(&url)
	if result.Error != nil {
		return nil
	}

	return url
}
