package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (app *application) indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "home.tmpl.html", app.newTemplateData(c))
}

func (app *application) aboutHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "about.tmpl.html", app.newTemplateData(c))
}

// signupHandler handles the display of the signup form.
func (app *application) signupHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "signup.tmpl.html", app.newTemplateData(c))
}
