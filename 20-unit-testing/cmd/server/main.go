package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"workshop/internal/bootstrap"
	"workshop/internal/router"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("error: running application: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	app, err := bootstrap.NewApp()
	if err != nil {
		return fmt.Errorf("error: initializing app: %w", err)
	}
	defer app.Cleanup()

	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", app.Config.Server.AppPort),
		Handler:      router.Api(app.Config, app.Database, app.Log, app.Validate),
		ReadTimeout:  app.Config.Server.ReadTimeout,
		WriteTimeout: app.Config.Server.WriteTimeout,
		IdleTimeout:  app.Config.Server.IdleTimeout,
	}

	serverErrChan := make(chan error, 1)

	// start server in a goroutine
	go func() {
		app.Log.Info(context.Background(), "starting server", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("error: listening and serving: %w", err)
		}
		close(serverErrChan)
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err, ok := <-serverErrChan:
		if ok && err != nil {
			app.Log.Error(context.Background(), "error: server error", slog.Any("error", err))
			return err
		}
	case sig := <-shutdownChan:
		app.Log.Info(context.Background(), "received shutdown signal", slog.String("signal", sig.String()))

		// Give more time for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), app.Config.Server.GracefulShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		if err := server.Shutdown(ctx); err != nil {
			app.Log.Error(context.Background(), "error during graceful shutdown", slog.Any("error", err))
			app.Log.Info(context.Background(), "attempting force close due to graceful shutdown failure")

			// Force close if graceful shutdown fails
			if err := server.Close(); err != nil && err != http.ErrServerClosed {
				app.Log.Error(context.Background(), "error during force close", slog.Any("error", err))
				return err
			}
		} else {
			app.Log.Info(context.Background(), "server gracefully shutdown complete")
		}
	}

	return nil
}
