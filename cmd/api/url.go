package main

import (
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
	fullUrl := c.Scheme() + "://" + c.Request().Host + "/s/" + url.ShortUrl
	return c.JSON(http.StatusCreated, model.UrlResponse{
		ID:      url.ID,
		FullUrl: fullUrl,
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
