package main

import (
	"fmt"
	"os"
)

type Config struct {
	host string
	user string
	pass string
}

type App struct {
	Config *Config
	Client *Client
}

func NewApp() *App {
	config := loadConfig()
	return &App{
		Config: config,
		Client: initClient(config),
	}
}

func loadConfig() *Config {
	return &Config{
		host: requireEnvVar("HOST"),
		user: requireEnvVar("USER"),
		pass: requireEnvVar("PASS"),
	}
}

func requireEnvVar(name string) string {
	value, found := os.LookupEnv(name)
	if !found {
		panic(fmt.Sprintf("environment variable %s not set", name))
	}
	return value
}

func initClient(config *Config) *Client {
	return NewClient(config.host)
}
