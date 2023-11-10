package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pascaldekloe/jwt"
)

func (app *application) authenticate(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if app.isAuthenticated(c) {
			return next(c)
		}
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

		c.Set("user", user)

		return next(c)
	}
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

		urlID := c.Param("id")
		urlUUID := uuid.MustParse(urlID)
		url := app.models.Urls.Find(urlUUID)

		if user.ID != url.UserID {
			return c.JSON(http.StatusUnauthorized, "Unauthorized")
		}

		return next(c)
	}
}
