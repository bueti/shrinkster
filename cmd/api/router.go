package main

import (
	"net/http"
	"strings"

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
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Path(), "/api")
		},
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
	app.echo.GET("/.well-known/security.txt", app.securityTxtHandler)

	// dashboard
	app.echo.GET("dashboard", app.dashboardHandler, app.authenticate)

	// user
	app.echo.GET("/users/activate", app.activateUserHandler)
	app.echo.GET("/users/resend-activation", app.resendActivationLinkHandler)
	app.echo.POST("/users/resend-activation", app.resendActivationLinkHandlerPost)
	app.echo.GET("/signup", app.signupHandler)
	app.echo.POST("/signup", app.signupHandlerPost)
	app.echo.GET("/login", app.loginHandler)
	app.echo.POST("/login", app.loginHandlerPost)
	app.echo.POST("/logout", app.logoutHandlerPost)

	// url
	app.echo.GET("/urls/new", app.createUrlFormHandler, app.authenticate)
	app.echo.POST("/urls", app.createUrlHandlerPost, app.authenticate)
	app.echo.POST("/urls/:id", app.deleteUrlHandlerPost, app.authenticate, app.mustBeOwner)
	app.echo.GET("/s/*", app.redirectUrlHandler)

	// create a group for all api calls. these accept json and return json
	api := app.echo.Group("/api")

	// healthcheck
	api.GET("/health", app.healthcheckHandlerJson)

	// api/users
	api.GET("/users", app.listUsersHandlerJson, app.authenticate, app.requireRole("admin"))
	api.GET("/users/:id", app.getUserHandlerJson, app.authenticate)
	api.GET("/users/activate", app.activateUserHandlerJson)
	api.POST("/users/resend-activation", app.resendActivationLinkHandlerJsonPost)
	api.POST("/signup", app.signupHandlerJsonPost)
	api.POST("/login", app.loginHandlerJsonPost)

	// api/urls
	api.POST("/urls", app.createUrlHandlerJsonPost, app.authenticate)
	api.POST("/urls/:id", app.deleteUrlHandlerJsonPost, app.authenticate, app.mustBeOwner)
	api.GET("/urls/:user_id", app.getUrlByUserHandlerJson, app.authenticate, app.mustBeOwner)
}
