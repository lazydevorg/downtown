package main

import (
	"fmt"
	"log/slog"
	"os"
)

type AppConfig struct {
	host string
	addr string
}

type App struct {
	Config *AppConfig
	Client *Client
	Logger *slog.Logger
}

func LoadAppConfig() *AppConfig {
	return &AppConfig{
		host: requireEnvVar("DOWNLOAD_STATION_HOST"),
		addr: optionalEnvVar("ADDR", ":4000"),
	}
}

func requireEnvVar(name string) string {
	value, found := os.LookupEnv(name)
	if !found {
		panic(fmt.Sprintf("environment variable %s not set", name))
	}
	return value
}

func optionalEnvVar(name string, defaultValue string) string {
	value, found := os.LookupEnv(name)
	if !found {
		return defaultValue
	}
	return value
}
