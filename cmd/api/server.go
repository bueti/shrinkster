package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
)

func (app *application) serve() error {
	// Start server
	go func() {
		addr := fmt.Sprintf(":%d", app.config.port)
		if err := app.echo.Start(addr); err != nil && err != http.ErrServerClosed {
			app.echo.Logger.Fatal("shutting down the server. error = %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.echo.Shutdown(ctx); err != nil {
		app.echo.Logger.Fatal(err)
	}

	return nil
}

func (app *application) indexHandler(c echo.Context) error {
	return c.Render(http.StatusOK, "base.tmpl.html", "Hello, World!")
}
