package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (app *application) indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "home.tmpl.html", "Hello, World!")
}

func (app *application) aboutHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "about.tmpl.html", "Hello, World!")
}

// signupHandler handles the submission of the signup form.
func (app *application) signupHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "signup.tmpl.html", "Hello, World!")
}
