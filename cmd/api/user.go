package main

import (
	"net/http"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) createUserHandler(c echo.Context) error {
	var body model.UserRegisterReq
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	userResponse, err := app.models.Users.Register(&body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
		//// why can I not use gorm.ErrDuplicatedKey here?
		//if resp.Error() == "ERROR: duplicate key value violates unique constraint \"idx_users_email\" (SQLSTATE 23505)" {
		//	return c.JSON(http.StatusBadRequest, "email already exists")
		//} else {
		//	return c.JSON(http.StatusInternalServerError, resp.Error())
		//}
	}

	return c.JSON(http.StatusCreated, userResponse)
}

// LoginUserHandler handles the login of a user
func (app *application) loginUserHandler(c echo.Context) error {
	var body model.UserLoginRequest
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user, err := app.models.Users.Login(body.Email, body.Password)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
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

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user, err := app.models.Users.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (app *application) listUsersHandler(c echo.Context) error {
	users, err := app.models.Users.List()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, users)
}

//func (app *application) updateUserHandler(c echo.Context) error {
//	user := new(model.User)
//	id := c.Param("id")
//	result := app.db.First(&user, id)
//	if result.Error != nil {
//		return c.JSON(http.StatusInternalServerError, result.Error.Error())
//	}
//	if err := c.Bind(user); err != nil {
//		return err
//	}
//	result = app.db.Save(&user)
//	if result.Error != nil {
//		return c.JSON(http.StatusInternalServerError, result.Error.Error())
//	}
//	return c.JSON(http.StatusOK, user)
//}
//
//func (app *application) deleteUserHandler(c echo.Context) error {
//	user := new(model.User)
//	id := c.Param("id")
//	result := app.db.First(&user, id)
//	if result.Error != nil {
//		return c.JSON(http.StatusInternalServerError, result.Error.Error())
//	}
//	result = app.db.Delete(&user)
//	if result.Error != nil {
//		return c.JSON(http.StatusInternalServerError, result.Error.Error())
//	}
//	return c.JSON(http.StatusOK, user)
//}
