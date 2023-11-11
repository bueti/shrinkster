package model

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Models struct {
	Urls   UrlModel
	Users  UserModel
	Roles  RoleModel
	Tokens TokenModel
}

func NewModels(db *gorm.DB) Models {
	return Models{
		Users:  UserModel{DB: db},
		Urls:   UrlModel{DB: db},
		Roles:  RoleModel{DB: db},
		Tokens: TokenModel{DB: db},
	}
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}

func base62Encode(number uint64) string {
	length := len(alphabet)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(alphabet[(number % uint64(length))])
	}

	return encodedBuilder.String()
}
