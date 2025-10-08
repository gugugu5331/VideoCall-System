package main

import (
	"fmt"
	"log"

	"meeting-system/shared/config"
)

func main() {
	cfg, err := config.LoadConfig("config/signaling-service.yaml")
	if err != nil {
		log.Fatalf("load err: %v", err)
	}
	fmt.Printf("secret=%q\n", cfg.JWT.Secret)
}
