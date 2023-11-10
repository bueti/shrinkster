package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (app *application) redirectUrlHandler(c echo.Context) error {
	wildcardValue := c.Param("*")
	shortUrl := strings.TrimSuffix(wildcardValue, "/")
	url, err := app.models.Urls.GetRedirect(shortUrl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.Redirect(http.StatusPermanentRedirect, url.Original)
}

func (app *application) createUrlHandler(c echo.Context) error {
	urlReq := new(model.UrlCreateRequest)
	if err := c.Bind(urlReq); err != nil {
		return err
	}

	url, err := app.models.Urls.Create(urlReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, model.UrlResponse{
		ID:      url.ID,
		FullUrl: genFullUrl(fmt.Sprintf(c.Scheme()+"://"+c.Request().Host), url.ShortUrl),
	})
}

func (app *application) getUrlByUserHandler(c echo.Context) error {
	userID := c.Param("user_id")
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	urls, err := app.models.Urls.GetUrlByUser(userUUID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, urls)
}

// deleteUrlHandler handles the deletion of a url.
func (app *application) deleteUrlHandler(c echo.Context) error {
	urlUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	err = app.models.Urls.Delete(urlUUID)
	if err != nil {
		return err
	}

	app.sessionManager.Put(c.Request().Context(), "flash", "Url deleted successfully!")
	data := app.newTemplateData(c)
	user, _ := app.userFromContext(c)
	data.User = user
	return c.Render(http.StatusCreated, "dashboard.tmpl.html", data)
}

// genFullUrl generates the full url for a given short url
func genFullUrl(prefix, url string) string {
	return prefix + "/s/" + url
}
