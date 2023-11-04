package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func initEcho() *echo.Echo {
	e := echo.New()
	e.Debug = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	return e
}

func (app *application) registerRoutes() {
	app.echo.GET("/health", app.healthcheckHandler)
	app.echo.GET("/users", app.listUsersHandler)
	app.echo.GET("/users/:id", app.getUserHandler)
	app.echo.POST("/users", app.createUserHandler)
	app.echo.PUT("/users/:id", app.updateUserHandler)
	app.echo.DELETE("/users/:id", app.deleteUserHandler)
}
