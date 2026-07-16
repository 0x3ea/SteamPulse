package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	httpadapter "github.com/0x3ea/SteamPulse/internal/adapter/httpapi"
	"github.com/0x3ea/SteamPulse/internal/core"
	"github.com/0x3ea/SteamPulse/internal/steam"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf(".env not loaded: %v (ok in production)", err)
	}

	// Wire the layers: steam client → core service → http handlers.
	steamClient, err := steam.New(os.Getenv("STEAM_API_KEY"))
	if err != nil {
		log.Fatalf("cannot start: %v — set STEAM_API_KEY in .env", err)
	}
	coreSvc := core.NewService(steamClient)
	handler := httpadapter.NewHandler(coreSvc)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	r := httpadapter.NewRouter(handler)
	log.Printf("SteamPulse listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
