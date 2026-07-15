package main

import (
	"log"
	"os"

	httpadapter "github.com/0x3ea/SteamPulse/internal/adapter/http"
)

func main() {
	if os.Getenv("STEAM_API_KEY") == "" {
		log.Println("warning: STEAM_API_KEY not set; Steam API calls will fail until it is")
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	r := httpadapter.NewRouter()
	log.Printf("SteamPulse listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
