package main

import (
	"net/http"
	"strings"

	"github.com/bueti/shrinkster/internal/model"
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
