package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserModel struct {
	DB *gorm.DB
}

type User struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Name      string    `gorm:"type:varchar(255)"`
	Email     string    `gorm:"not null;uniqueIndex"`
	Password  string    `gorm:"not null"`
	Role      string    `gorm:"default:'user'"`
	Activated bool      `gorm:"default:false"`
}

type UserRegisterReq struct {
	Email           string `json:"email" validate:"required,email"`
	Name            string `json:"name" validate:"required"`
	Password        string `json:"password" validate:"required,min=8,max=72"`
	PasswordConfirm string `json:"password_confirm" validate:"required,min=8,max=72"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type UserLoginResponse struct {
	ID    uuid.UUID `json:"id"`
	Token string    `json:"token"`
}

func (u *UserModel) Login(email, password string) (*User, error) {
	user := new(User)
	result := u.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	if !checkPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}
	return user, nil
}

func (u *UserModel) Register(body *UserRegisterReq) (UserResponse, error) {
	hashedPassword, err := hashPassword(body.Password)
	if err != nil {
		return UserResponse{}, err
	}

	user := &User{
		Email:    body.Email,
		Name:     body.Name,
		Password: hashedPassword,
	}
	result := u.DB.Create(&user)
	if result.Error != nil {
		return UserResponse{}, result.Error
	}

	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Activate sets the activated flag to true for a user.
func (u *UserModel) Activate(id uuid.UUID) error {
	user, err := u.GetByID(id)
	if err != nil {
		return err
	}
	user.Activated = true
	result := u.DB.Save(&user)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (u *UserModel) List() ([]User, error) {
	var users []User
	result := u.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (u *UserModel) GetByID(id uuid.UUID) (*User, error) {
	user := new(User)
	result := u.DB.First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return user, nil
}

func (u *UserModel) GetRole(c echo.Context) (string, error) {
	user := c.Get("user").(*User)
	return user.Role, nil
}

// checkPasswordHash compares a plain text password with a hashed password
// and returns true if they match or false otherwise.
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}
	return true
}
