package main

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/labstack/echo/v4"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func (app *application) redirectUrlHandler(c echo.Context) error {
	url := new(model.Url)

	wildcardValue := c.Param("*")
	shortUrl := strings.TrimSuffix(wildcardValue, "/")
	result := app.db.Where("short_url = ?", shortUrl).First(&url)

	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.Redirect(http.StatusPermanentRedirect, url.Original)
}

func (app *application) listUrlsHandler(c echo.Context) error {
	var urls []model.Url
	result := app.db.Find(&urls)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, urls)
}

func (app *application) createUrlHandler(c echo.Context) error {
	url := new(model.Url)
	if err := c.Bind(url); err != nil {
		return err
	}

	if url.ShortCode != "" {
		url.ShortUrl = url.ShortCode
	} else {
		id := base62Encode(rand.Uint64())
		url.ShortUrl = id
	}

	resp := app.db.Create(url)
	if resp.Error != nil {
		return c.JSON(http.StatusInternalServerError, resp.Error.Error())
	}
	fullUrl := c.Scheme() + "://" + c.Request().Host + "/s/" + url.ShortUrl
	return c.JSON(http.StatusCreated, fullUrl)
}

func (app *application) updateUrlHandler(c echo.Context) error {
	url := new(model.Url)
	id := c.Param("id")
	result := app.db.First(&url, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	if err := c.Bind(url); err != nil {
		return err
	}
	result = app.db.Save(&url)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, url)
}

func (app *application) deleteUrlHandler(c echo.Context) error {
	url := new(model.Url)
	id := c.Param("id")
	result := app.db.First(&url, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	result = app.db.Delete(&url)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, url)
}

func base62Encode(number uint64) string {
	length := len(alphabet)
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(10)
	for ; number > 0; number = number / uint64(length) {
		encodedBuilder.WriteByte(alphabet[(number % uint64(length))])
	}

	return encodedBuilder.String()
}
