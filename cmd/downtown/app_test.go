package main

import (
	"testing"
)

func TestLoadAppConfig(t *testing.T) {
	t.Setenv("DOWNLOAD_STATION_HOST", "DSHOST")

	config := LoadAppConfig()
	if config.host != "DSHOST" {
		t.Errorf("config.host = %s; want DSHOST", config.host)
	}
	if config.addr != ":4000" {
		t.Errorf("config.addr = %s; want default value :4000", config.addr)
	}

	t.Setenv("ADDR", "localhost:8000")
	config = LoadAppConfig()
	if config.addr != "localhost:8000" {
		t.Errorf("config.addr = %s; want value localhost:8000", config.addr)
	}
}

func TestLoadAppConfigPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Panic expected when environment variable DOWNLOAD_STATION_HOST is not set")
		}
	}()
	_ = LoadAppConfig()
}
