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

// signupHandler handles the display of the signup form.
func (app *application) signupHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "signup.tmpl.html", app.newTemplateData(c))
}

func (app *application) signupHandlerPost(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	passwordConfirm := c.FormValue("password_confirm")

	emailRX := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
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

	newUser, _ := app.models.Users.GetByID(user.ID)
	sendActivationEmail(token, newUser, app)

	app.sessionManager.Put(c.Request().Context(), "flash", "Your signup was successful. Please check your mailbox for the account activation link.")
	return c.Redirect(http.StatusSeeOther, "/login")
}

func (app *application) signupHandlerJsonPost(c echo.Context) error {
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

func (app *application) loginHandlerPost(c echo.Context) error {
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

func (app *application) loginHandlerJsonPost(c echo.Context) error {
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

func (app *application) getUserHandlerJson(c echo.Context) error {

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

	userID, err := app.models.Tokens.GetUserID(model.ScopeActivation, token)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Invalid token or token expired.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusBadRequest, "home.tmpl.html", data)
	}

	err = app.models.Users.Activate(userID)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "home.tmpl.html", app.newTemplateData(c))
	}
	app.sessionManager.Put(c.Request().Context(), "flash", "Your account has been activated successfully. Please log in.")
	data := app.newTemplateData(c)
	return c.Render(http.StatusOK, "home.tmpl.html", data)
}

// activateUserHandlerJson handles the activation of a user with json.
func (app *application) activateUserHandlerJson(c echo.Context) error {
	req := struct {
		Token string `json:"token"`
	}{}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	err := model.ValidateTokenPlaintext(req.Token)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	userID, err := app.models.Tokens.GetUserID(model.ScopeActivation, req.Token)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = app.models.Users.Activate(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "Your account has been activated successfully. Please log in.")
}

// resendActivationLinkHandler handles the display of the resend activation link form.
func (app *application) resendActivationLinkHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "resend_activation_link.tmpl.html", app.newTemplateData(c))
}

// resendActivationLinkHandlerPost handles the resending of an activation link.
func (app *application) resendActivationLinkHandlerPost(c echo.Context) error {
	email := c.FormValue("email")
	user, err := app.models.Users.GetByEmail(email)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "No user exists with this email address.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusBadRequest, "home.tmpl.html", data)
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, model.ScopeActivation)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Internal Server Error. Please try again later.")
		data := app.newTemplateData(c)
		return c.Render(http.StatusBadRequest, "home.tmpl.html", data)
	}

	sendActivationEmail(token, user, app)

	app.sessionManager.Put(c.Request().Context(), "flash", "The activation link has been resent. Please check your mailbox.")
	data := app.newTemplateData(c)
	return c.Render(http.StatusOK, "home.tmpl.html", data)
}

// resendActivationLinkHandlerJsonPost handles the resending of an activation link with json.
func (app *application) resendActivationLinkHandlerJsonPost(c echo.Context) error {
	body := struct {
		Email string `json:"email"`
	}{}

	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user, err := app.models.Users.GetByEmail(body.Email)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, model.ScopeActivation)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	sendActivationEmail(token, user, app)

	return c.JSON(http.StatusOK, "The activation link has been resent. Please check your mailbox.")
}

func (app *application) listUsersHandlerJson(c echo.Context) error {
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

// logoutHandlerPost handles the logout of a user.
func (app *application) logoutHandlerPost(c echo.Context) error {
	err := app.sessionManager.RenewToken(c.Request().Context())
	if err != nil {
		return c.Render(http.StatusInternalServerError, "home.tmpl.html", app.newTemplateData(c))
	}

	app.sessionManager.Remove(c.Request().Context(), "authenticated")
	c.Set("user", nil)
	app.sessionManager.Put(c.Request().Context(), "flash", "You've been logged out successfully!")
	return c.Render(http.StatusOK, "home.tmpl.html", app.newTemplateData(c))
}

func sendActivationEmail(token *model.Token, user *model.User, app *application) {
	go func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		err := app.mailer.Send(user.Email, "welcome.tmpl.html", data)
		if err != nil {
			log.Error(err)
		}
	}()
}
