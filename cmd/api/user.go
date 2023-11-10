package main

import (
	"net/http"
	"time"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pascaldekloe/jwt"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) createUserHandler(c echo.Context) error {
	contentType := c.Request().Header.Get(echo.HeaderContentType)
	switch contentType {
	case echo.MIMEApplicationJSON:
		return app.handleJSONSignup(c)
	case echo.MIMEApplicationForm:
		return app.handleFormSignup(c)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unsupported content type"})
	}
}

func (app *application) handleFormSignup(c echo.Context) error {
	user, err := app.models.Users.Register(&model.UserRegisterReq{
		Name:            c.FormValue("name"),
		Email:           c.FormValue("email"),
		Password:        c.FormValue("password"),
		PasswordConfirm: c.FormValue("password_confirm"),
	})
	if err != nil {
		return c.Render(http.StatusBadRequest, "signup.tmpl.html", map[string]interface{}{
			"Error": err.Error(),
		})
	}

	app.sessionManager.Put(c.Request().Context(), "authenticated", "true")
	app.sessionManager.Put(c.Request().Context(), "flash", "registered successfully")
	c.Set("user", user)

	return c.Redirect(http.StatusSeeOther, "/")
}

func (app *application) handleJSONSignup(c echo.Context) error {
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
	contentType := c.Request().Header.Get(echo.HeaderContentType)
	switch contentType {
	case echo.MIMEApplicationJSON:
		return app.handleJSONLogin(c)
	case echo.MIMEApplicationForm:
		return app.handleFormLogin(c)
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Unsupported content type"})
	}
}

func (app *application) handleFormLogin(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	user, err := app.models.Users.Login(email, password)
	if err != nil {
		return c.Render(http.StatusUnauthorized, "login.tmpl.html", map[string]interface{}{
			"Error": "Invalid credentials",
		})
	}

	app.sessionManager.Put(c.Request().Context(), "authenticated", "true")
	app.sessionManager.Put(c.Request().Context(), "flash", "logged in successfully")

	c.Set("user", user)

	return c.Redirect(http.StatusSeeOther, "/")

}

func (app *application) handleJSONLogin(c echo.Context) error {
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

	var claims jwt.Claims
	claims.Subject = user.ID.String()
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "shrink.ch"
	claims.Audiences = []string{"shrink.ch"}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(app.config.signingKey))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	userLoginResponse := model.UserLoginResponse{
		ID:    user.ID,
		Token: string(jwtBytes),
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

// loginHandler handles the display of the login form.
func (app *application) loginHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "login.tmpl.html", nil)
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
