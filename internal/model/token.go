package model

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

// ValidateTokenPlaintext validates a token plaintext.
// token must not be empty and be 26 bytes long.
func ValidateTokenPlaintext(tokenPlaintext string) error {
	if len(tokenPlaintext) != 26 {
		return fmt.Errorf("token must be 26 bytes long")
	}
	return nil
}

type Token struct {
	gorm.Model
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    uuid.UUID `json:"-"`
	User      User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userID uuid.UUID, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

type TokenModel struct {
	DB *gorm.DB
}

func (m TokenModel) New(userID uuid.UUID, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	result := m.DB.Create(token)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (m TokenModel) DeleteAllForUser(scope string, userID uuid.UUID) error {
	result := m.DB.Where("scope = ? AND user_id = ?", scope, userID).Delete(&Token{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// GetUser returns a user for a given token.
func (m TokenModel) GetUser(scope, token string) (*User, error) {
	tokenObj := new(Token)
	tokenHash := sha256.Sum256([]byte(token))
	result := m.DB.Where("scope = ? AND hash = ?", scope, tokenHash[:]).First(&tokenObj)
	if result.Error != nil {
		return nil, result.Error
	}

	return &tokenObj.User, nil
}
