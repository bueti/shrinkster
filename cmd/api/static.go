package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (app *application) indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "home.tmpl.html", nil)
}

func (app *application) aboutHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "about.tmpl.html", nil)
}

// signupHandler handles the display of the signup form.
func (app *application) signupHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "signup.tmpl.html", nil)
}

// loginHandler handles the display of the login form.
func (app *application) loginHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "login.tmpl.html", nil)
}
