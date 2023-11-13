package main

import (
	"net/http"
	"regexp"
	"time"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
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
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")

	emailRX := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRX.MatchString(email) {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Invalid email address.")
		return c.Render(http.StatusBadRequest, "login.tmpl.html", app.newTemplateData(c))
	}

	if len(password) < 8 || len(password) > 72 {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Password must be between 8 and 72 characters long.")
		return c.Render(http.StatusBadRequest, "signup.tmpl.html", app.newTemplateData(c))
	}

	if password != passwordConfirm {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Password does not match.")
		return c.Render(http.StatusBadRequest, "signup.tmpl.html", app.newTemplateData(c))
	}

	user, err := app.models.Users.Register(&model.UserRegisterReq{
		Name:     name,
		Email:    email,
		Password: password,
	})
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Internal Server Error. Please try again later.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusBadRequest, "signup.tmpl.html", data)
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, model.ScopeActivation)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Internal Server Error. Please try again later.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusBadRequest, "signup.tmpl.html", data)
	}

	// Launch a goroutine which runs an anonymous function that sends the welcome email.

	go func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err = app.mailer.Send(user.Email, "welcome.tmpl.html", data)
		if err != nil {
			log.Error(err)
		}
	}()

	app.sessionManager.Put(c.Request().Context(), "flash", "Your signup was successful. Please check your mailbox for the account activation link.")
	return c.Redirect(http.StatusSeeOther, "/login")
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
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Login failed. Please check your username and password and try again.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusUnauthorized, "login.tmpl.html", data)
	}
	if !user.Activated {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Your user account has not been activated. Please check your mailbox for the activation link.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusUnauthorized, "login.tmpl.html", data)
	}
	userID := user.ID.String()
	app.sessionManager.Put(c.Request().Context(), "authenticated", true)
	app.sessionManager.Put(c.Request().Context(), "userID", userID)
	app.sessionManager.Put(c.Request().Context(), "flash", "Logged in successfully")

	data := app.newTemplateData(c)
	return c.Render(http.StatusOK, "home.tmpl.html", data)
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

// activateUserHandler handles the activation of a user.
func (app *application) activateUserHandler(c echo.Context) error {
	token := c.QueryParam("token")
	err := model.ValidateTokenPlaintext(token)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	user, err := app.models.Tokens.GetUser(model.ScopeActivation, token)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Invalid token or token expired.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusBadRequest, "home.tmpl.html", data)
	}

	err = app.models.Users.Activate(user.ID)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "home.tmpl.html", app.newTemplateData(c))
	}
	app.sessionManager.Put(c.Request().Context(), "flash", "Your account has been activated successfully. Please log in.")
	data := app.newTemplateData(c)
	return c.Render(http.StatusOK, "home.tmpl.html", data)
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
	return c.Render(http.StatusOK, "login.tmpl.html", app.newTemplateData(c))
}

// logoutHandler handles the logout of a user.
func (app *application) logoutHandler(c echo.Context) error {
	err := app.sessionManager.RenewToken(c.Request().Context())
	if err != nil {
		return c.Render(http.StatusInternalServerError, "home.tmpl.html", app.newTemplateData(c))
	}

	app.sessionManager.Remove(c.Request().Context(), "authenticated")
	c.Set("user", nil)
	app.sessionManager.Put(c.Request().Context(), "flash", "You've been logged out successfully!")
	return c.Render(http.StatusOK, "home.tmpl.html", app.newTemplateData(c))
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
