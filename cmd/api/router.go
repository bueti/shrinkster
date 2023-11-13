package main

import (
	"net/http"

	"github.com/bueti/shrinkster/ui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spazzymoto/echo-scs-session"
)

func (app *application) initEcho() *echo.Echo {
	e := echo.New()

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
	app.echo.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "form:csrf_token",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,
	}))
	app.echo.Use(middleware.Secure())
	app.echo.Use(middleware.BodyLimit("1M"))
	app.echo.Use(middleware.RequestID())
	app.echo.Use(session.LoadAndSave(app.sessionManager))
}

func (app *application) registerRoutes() {
	fileServer := http.FileServer(http.FS(ui.Files))
	app.echo.GET("/static/*filepath", echo.WrapHandler(fileServer))

	// static pages
	app.echo.GET("/", app.indexHandler)
	app.echo.GET("/about", app.aboutHandler)

	// healthcheck
	app.echo.GET("/health", app.healthcheckHandler)
	app.echo.GET("/.well-known/security.txt", app.securityTxtHandler)

	// dashboard
	app.echo.GET("dashboard", app.dashboardHandler, app.authenticate)

	// user
	app.echo.GET("/users", app.listUsersHandler, app.authenticate, app.requireRole("admin"))
	app.echo.GET("/users/:id", app.getUserHandler, app.authenticate)
	app.echo.GET("/users/activate", app.activateUserHandler)
	app.echo.GET("/signup", app.signupHandler)
	app.echo.POST("/signup", app.createUserHandler)
	app.echo.GET("/login", app.loginHandler)
	app.echo.POST("/login", app.loginUserHandler)
	app.echo.POST("/logout", app.logoutHandler)

	// url
	app.echo.GET("/urls/new", app.createUrlFormHandler, app.authenticate)
	app.echo.POST("/urls", app.createUrlHandler, app.authenticate)
	app.echo.POST("/urls/:id", app.deleteUrlHandler, app.authenticate, app.mustBeOwner)
	app.echo.GET("/urls/:user_id", app.getUrlByUserHandler, app.authenticate, app.mustBeOwner)
	app.echo.GET("/s/*", app.redirectUrlHandler)
}
