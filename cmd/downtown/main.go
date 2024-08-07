package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	appConfig := LoadAppConfig()

	var logLevel slog.Level
	if appConfig.devMode == "true" {
		logLevel = slog.LevelDebug
	} else {
		logLevel = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	client := NewClient(appConfig.host, logger)

	app := App{
		Config: appConfig,
		Client: client,
		Logger: logger,
	}

	webapp := WebApp{
		App:       &app,
		Logger:    logger,
		Templates: LoadTemplates(),
	}
	srv := &http.Server{
		Addr:         appConfig.addr,
		Handler:      webapp.routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  time.Minute,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("Server started", "addr", srv.Addr)
	err := srv.ListenAndServe()
	if err != nil {
		logger.Error("Shutting down the server: %v", err)
		os.Exit(1)
	}
}
