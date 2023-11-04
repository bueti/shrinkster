package main

import (
	"github.com/labstack/echo-contrib/echoprometheus"
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
	e.Use(echoprometheus.NewMiddleware("shrinkster"))
	e.GET("/metrics", echoprometheus.NewHandler())

	return e
}

func (app *application) registerRoutes() {
	// healthcheck
	app.echo.GET("/health", app.healthcheckHandler)
	// user routes
	app.echo.GET("/users", app.listUsersHandler)
	app.echo.GET("/users/:id", app.getUserHandler)
	app.echo.POST("/users", app.createUserHandler)
	app.echo.PUT("/users/:id", app.updateUserHandler)
	app.echo.DELETE("/users/:id", app.deleteUserHandler)
	// url routes
	app.echo.GET("/urls", app.listUrlsHandler)
	app.echo.POST("/urls", app.createUrlHandler)
	app.echo.PUT("/urls/:id", app.updateUrlHandler)
	app.echo.DELETE("/urls/:id", app.deleteUrlHandler)
	app.echo.GET("/s/*", app.redirectUrlHandler)
}
