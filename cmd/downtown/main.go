package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	appConfig := LoadAppConfig()
	client := NewClient(appConfig.host)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := App{
		Config: appConfig,
		Client: client,
		Logger: logger,
	}

	webapp := WebApp{
		App: &app,
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
