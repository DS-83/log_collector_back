package main

import (
	"log"
	"log_collect/config"
	"log_collect/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	app := server.NewApp(cfg)

	if err := app.Run(cfg); err != nil {
		log.Fatalf("%s", err.Error())
	}
}
