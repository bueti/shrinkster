package main

import (
	"fmt"
	"net/http"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/bueti/shrinkster/ui"
	"github.com/google/uuid"
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

// securityTxtHandler handles the display of the security.txt file.
func (app *application) securityTxtHandler(c echo.Context) error {
	content, err := ui.Files.ReadFile("static/security.txt")
	if err != nil {
		return c.String(http.StatusInternalServerError, "Internal Server Error")
	}
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlainCharsetUTF8)
	return c.String(http.StatusOK, string(content))
}

// dashboardHandler handles the display of the dashboard page.
func (app *application) dashboardHandler(c echo.Context) error {
	user, err := app.userFromContext(c)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Bad Request, are you logged in?")
		return c.Render(http.StatusInternalServerError, "home.tmpl.html", app.newTemplateData(c))
	}

	data := app.newTemplateData(c)
	urlsResp, err := app.models.Urls.GetUrlByUser(user.ID)
	if err != nil {
		app.sessionManager.Put(c.Request().Context(), "flash_error", "Internal Server Error. Please try again later.")
		return c.Render(http.StatusInternalServerError, "dashboard.tmpl.html", data)
	}
	var urls []*model.Url
	for _, urlByUserResponse := range *urlsResp {
		var url model.Url
		url.ID = urlByUserResponse.ID
		url.ShortUrl = genFullUrl(fmt.Sprintf(c.Scheme()+"://"+c.Request().Host), urlByUserResponse.ShortUrl)
		url.Original = urlByUserResponse.Original
		url.Visits = urlByUserResponse.Visits
		url.CreatedAt = urlByUserResponse.CreatedAt
		url.UpdatedAt = urlByUserResponse.UpdatedAt

		urls = append(urls, &url)
	}
	data.Urls = urls
	data.User = user
	return c.Render(http.StatusOK, "dashboard.tmpl.html", data)
}

func (app *application) userFromContext(c echo.Context) (*model.User, error) {
	userID := app.sessionManager.Get(c.Request().Context(), "userID")
	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		return nil, err
	}
	user, err := app.models.Users.GetByID(userUUID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
