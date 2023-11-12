package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func (app *application) serve() error {
	// Start server
	go func() {
		// Retrieve TLS key and certificate content from environment variables
		tlsKeyContent := os.Getenv("TLS_KEY")
		tlsCertContent := os.Getenv("TLS_CERT")

		addr := fmt.Sprintf(":%d", app.config.port)
		if err := app.echo.StartTLS(addr, []byte(tlsCertContent), []byte(tlsKeyContent)); err != nil && err != http.ErrServerClosed {
			app.echo.Logger.Fatal("shutting down the server. error = %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
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
