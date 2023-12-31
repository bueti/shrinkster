package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/bueti/shrinkster/internal/model"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pascaldekloe/jwt"
)

func (app *application) authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get(echo.HeaderContentType) == echo.MIMEApplicationJSON {
			return app.jsonAuthenticate(c, next)
		}
		if !app.isAuthenticated(c) {
			return c.Render(http.StatusUnauthorized, "login.tmpl.html", app.newTemplateData(c))
		}
		c.Request().Header.Set("Cache-Control", "no-store")
		return next(c)

	}
}

func (app *application) jsonAuthenticate(c echo.Context, next echo.HandlerFunc) error {
	authorizationHeader := c.Request().Header.Get("Authorization")
	if authorizationHeader == "" {
		return c.JSON(http.StatusUnauthorized, "Unauthorized")
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return c.JSON(http.StatusBadRequest, "Bad Request")
	}

	token := headerParts[1]
	claims, err := jwt.HMACCheck([]byte(token), []byte(app.config.signingKey))
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid Token")
	}
	if !claims.Valid(time.Now()) {
		return c.JSON(http.StatusBadRequest, "Invalid Token")
	}
	if claims.Issuer != "shrink.ch" {
		return c.JSON(http.StatusBadRequest, "Invalid Token")
	}
	if !claims.AcceptAudience("shrink.ch") {
		return c.JSON(http.StatusBadRequest, "Invalid Token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid Token")
	}

	user, err := app.models.Users.GetByID(userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid Token")
	}

	app.sessionManager.Put(c.Request().Context(), "userID", user.ID.String())

	return next(c)
}

func (app *application) requireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the authenticated user's role
			userRole, err := app.models.Users.GetRole(c)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, "Unauthorized")
			}

			// Check if the user has the required role
			if userRole != role {
				return c.JSON(http.StatusForbidden, "Access Denied")
			}

			// If the user has the required role, call the next handler
			return next(c)
		}
	}
}

func (app *application) mustBeOwner(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := app.userFromContext(c)
		if err != nil {
			return c.Render(http.StatusUnauthorized, "home.tmpl.html", app.newTemplateData(c))
		}
		// admins are allowed to see anything
		if user.Role == "admin" {
			return next(c)
		}

		handlerName := c.Path()
		if handlerName == "/api/urls" && c.Request().Method == http.MethodDelete {
			urlReq := new(model.UrlDeleteRequest)
			if err := c.Bind(urlReq); err != nil {
				return c.JSON(http.StatusBadRequest, err.Error())
			}
			// Store urlReq in context
			app.sessionManager.Put(c.Request().Context(), "urlReq", urlReq)
			//c.Set("urlReq", urlReq)

			url := app.models.Urls.Find(urlReq.ID)

			if url.UserID != user.ID {
				return c.JSON(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}
		if handlerName == "/api/urls/:user_id" {
			userReqID := c.Param("user_id")
			userRedUUID, err := uuid.Parse(userReqID)
			if err != nil {
				return c.JSON(http.StatusBadRequest, err.Error())
			}

			if user.ID != userRedUUID {
				return c.JSON(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}
		if handlerName == "/urls/:id" && c.Request().Method == http.MethodPost {
			urlUUID, err := uuid.Parse(c.Param("id"))
			if err != nil {
				return c.JSON(http.StatusBadRequest, err.Error())
			}

			url := app.models.Urls.Find(urlUUID)

			if url.UserID != user.ID {
				return c.JSON(http.StatusUnauthorized, "Unauthorized")
			}
			return next(c)
		}

		return c.JSON(http.StatusUnauthorized, "Unauthorized")
	}
}
