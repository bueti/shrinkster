package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) createUserHandler(c echo.Context) error {
	var body model.UserRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	emailRX := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRX.MatchString(body.Email) {
		return c.JSON(http.StatusBadRequest, "invalid email address")
	}

	if body.Password != body.PasswordConfirm {
		return c.JSON(http.StatusBadRequest, "passwords do not match")
	}

	hashedPassword, err := hashPassword(body.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	user := &model.User{
		Email:    body.Email,
		Name:     body.Name,
		Password: hashedPassword,
	}

	resp := app.db.Create(user)
	if resp.Error != nil {
		// why can I not use gorm.ErrDuplicatedKey here?
		if resp.Error.Error() == "ERROR: duplicate key value violates unique constraint \"idx_users_email\" (SQLSTATE 23505)" {
			return c.JSON(http.StatusBadRequest, "email already exists")
		} else {
			return c.JSON(http.StatusInternalServerError, resp.Error.Error())
		}
	}

	userResponse := model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.JSON(http.StatusCreated, userResponse)
}

// LoginUserHandler handles the login of a user
func (app *application) loginUserHandler(c echo.Context) error {
	var body model.UserLoginRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user := &model.User{}
	result := app.db.Where("email = ?", body.Email).First(&user)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, "invalid credentials")
	}

	token, err := jwt.New(jwt.SigningMethodHS256).SignedString([]byte(app.config.signingKey))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	userLoginResponse := model.UserLoginResponse{
		ID:    user.ID,
		Token: token,
	}

	return c.JSON(http.StatusOK, userLoginResponse)
}

func (app *application) getUserHandler(c echo.Context) error {
	user := new(model.User)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	result := app.db.First(&user, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}

	userResponse := model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.JSON(http.StatusOK, userResponse)
}

func (app *application) listUsersHandler(c echo.Context) error {
	var users []model.User
	result := app.db.Find(&users)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, users)
}

func (app *application) updateUserHandler(c echo.Context) error {
	user := new(model.User)
	id := c.Param("id")
	result := app.db.First(&user, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	if err := c.Bind(user); err != nil {
		return err
	}
	result = app.db.Save(&user)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (app *application) deleteUserHandler(c echo.Context) error {
	user := new(model.User)
	id := c.Param("id")
	result := app.db.First(&user, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	result = app.db.Delete(&user)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", fmt.Errorf("could not hash password %w", err)
	}
	return string(hashedPassword), nil
}
