package main

import (
	"net/http"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/labstack/echo/v4"
)

func (app *application) createUserHandler(c echo.Context) error {
	user := new(model.User)
	if err := c.Bind(user); err != nil {
		return err
	}
	resp := app.db.Create(user)
	if resp.Error != nil {
		return c.JSON(http.StatusInternalServerError, resp.Error.Error())
	}
	return c.JSON(http.StatusCreated, user)
}

func (app *application) getUserHandler(c echo.Context) error {
	user := new(model.User)
	id := c.Param("id")
	result := app.db.First(&user, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (app *application) listUsersHandler(c echo.Context) error {
	users := []model.User{}
	result := app.db.Find(&users)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, users)
}

func (app *application) updateUserHandler(c echo.Context) error {
	user := new(model.User)
	id := c.Param("id")
	result := app.db.First(&user, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	if err := c.Bind(user); err != nil {
		return err
	}
	result = app.db.Save(&user)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func (app *application) deleteUserHandler(c echo.Context) error {
	user := new(model.User)
	id := c.Param("id")
	result := app.db.First(&user, id)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	result = app.db.Delete(&user)
	if result.Error != nil {
		return c.JSON(http.StatusInternalServerError, result.Error.Error())
	}
	return c.JSON(http.StatusOK, user)
}
