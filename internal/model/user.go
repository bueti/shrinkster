package model

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	DB *gorm.DB
}

type User struct {
	gorm.Model
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primary_key"`
	Name     string    `gorm:"type:varchar(255)"`
	Email    string    `gorm:"not null;uniqueIndex"`
	Password string    `gorm:"not null"`
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
	return user, nil
}

func (u *UserModel) Register(body *UserRegisterReq) (UserResponse, error) {
	emailRX := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRX.MatchString(body.Email) {
		//return c.JSON(http.StatusBadRequest, "invalid email address")
		return UserResponse{}, fmt.Errorf("invalid email address")
	}

	if body.Password != body.PasswordConfirm {
		//return c.JSON(http.StatusBadRequest, "passwords do not match")
		return UserResponse{}, fmt.Errorf("passwords do not match")
	}

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
