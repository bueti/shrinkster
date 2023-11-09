package main

import (
	"net/http"

	"github.com/bueti/shrinkster/ui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func (app *application) initEcho() *echo.Echo {
	e := echo.New()
	e.Debug = true
	e.Logger.SetLevel(log.DEBUG)

	e.Renderer = &Template{
		templates: app.initTemplate(),
	}

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
}

func (app *application) registerRoutes() {
	fileServer := http.FileServer(http.FS(ui.Files))
	app.echo.GET("/static/*filepath", echo.WrapHandler(fileServer))

	app.echo.GET("/", app.indexHandler)
	app.echo.GET("/about", app.aboutHandler)
	app.echo.GET("/signup", app.signupHandler)
	//app.echo.GET("/login", app.loginHandler)
	//app.echo.GET("/logout", app.logoutHandler)

	// healthcheck
	app.echo.GET("/health", app.healthcheckHandler)
	// user routes
	app.echo.GET("/users", app.listUsersHandler, app.authenticate, app.requireRole("admin"))
	app.echo.GET("/users/:id", app.getUserHandler, app.authenticate)
	//app.echo.PUT("/users/:id", app.updateUserHandler)
	//app.echo.DELETE("/users/:id", app.deleteUserHandler)
	app.echo.POST("/signup", app.createUserHandler)
	app.echo.POST("/login", app.loginUserHandler)
	// url routes
	app.echo.POST("/urls", app.createUrlHandler, app.authenticate)
	app.echo.GET("/urls/:user_id", app.getUrlByUserHandler, app.authenticate, app.mustBeOwner)
	app.echo.GET("/s/*", app.redirectUrlHandler)
}
