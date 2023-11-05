package main

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func initEcho() *echo.Echo {
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	return e
}

func (app *application) registerMiddleware() {
	app.echo.Use(middleware.Logger())
	app.echo.Use(middleware.Recover())
	app.echo.Use(middleware.Gzip())
	app.echo.Use(middleware.CORS())
	app.echo.Use(middleware.Secure())
	app.echo.Use(middleware.BodyLimit("1M"))
	app.echo.Use(middleware.RequestID())
	app.echo.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey: []byte(app.config.signingKey),
		Skipper: func(c echo.Context) bool {
			// Skip authentication for some paths
			if c.Path() == "/login" || c.Path() == "/signup" || c.Path() == "/health" || c.Path() == "/s/*" {
				return true
			}
			return false
		},
	}))
}

func (app *application) registerRoutes() {
	// healthcheck
	app.echo.GET("/health", app.healthcheckHandler)
	// user routes
	app.echo.GET("/users", app.listUsersHandler)
	app.echo.GET("/users/:id", app.getUserHandler)
	//app.echo.PUT("/users/:id", app.updateUserHandler)
	//app.echo.DELETE("/users/:id", app.deleteUserHandler)
	app.echo.POST("/signup", app.createUserHandler)
	app.echo.POST("/login", app.loginUserHandler)
	// url routes
	app.echo.POST("/urls", app.createUrlHandler)
	app.echo.GET("/urls/:user_id", app.getUrlByUserHandler)
	app.echo.GET("/s/*", app.redirectUrlHandler)
}
