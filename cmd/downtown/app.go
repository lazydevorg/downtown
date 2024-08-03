package main

import (
	"fmt"
	"log/slog"
	"os"
)

type AppConfig struct {
	host string
	user string
	pass string
	addr string
}

type App struct {
	Config *AppConfig
	Client *Client
	Logger *slog.Logger
}

func LoadAppConfig() *AppConfig {
	return &AppConfig{
		host: requireEnvVar("HOST"),
		user: requireEnvVar("USER"),
		pass: requireEnvVar("PASS"),
		addr: requireEnvVar("ADDR"),
	}
}

func requireEnvVar(name string) string {
	value, found := os.LookupEnv(name)
	if !found {
		panic(fmt.Sprintf("environment variable %s not set", name))
	}
	return value
}
