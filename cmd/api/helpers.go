package main

import (
	"time"

	"github.com/labstack/echo/v4"
)

func (app *application) isAuthenticated(c echo.Context) bool {
	return app.sessionManager.GetBool(c.Request().Context(), "authenticated")
}

func (app *application) newTemplateData(c echo.Context) *templateData {
	return &templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(c.Request().Context(), "flash"),
		FlashError:      app.sessionManager.PopString(c.Request().Context(), "flash_error"),
		IsAuthenticated: app.isAuthenticated(c),
	}
}
