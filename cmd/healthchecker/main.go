package main

import (
	"github.com/alexflint/go-arg"
	"net/http"
	"os"
)

type Config struct {
	ServicePort string `arg:"env:SERVICE_PORT" default:"6767" help:"port on localhost to check"`
}

func main() {
	// Initialize config struct and populate it from env vars and flags.
    cfg := &Config{}
    arg.MustParse(cfg)

	resp, err := http.Get("http://127.0.0.1:" + cfg.ServicePort + "/health") // Note pointer dereference using *

	// If there is an error or non-200 status, exit with 1 signaling unsuccessful check.
	if err != nil || resp.StatusCode != 200 {
		os.Exit(1)
	}
	os.Exit(0)
}
