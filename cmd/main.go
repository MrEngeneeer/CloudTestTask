package main

import (
	"balancer/internal/config"
	"balancer/internal/server"
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "configs/config.yml", "path for config")
	flag.Parse()
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	if err := server.Run(cfg); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
