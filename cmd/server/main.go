package main

import (
	"fmt"
	"log"

	"greenlabelai/backend/internal/api"
	"greenlabelai/backend/internal/config"
	"greenlabelai/backend/internal/db"
)

func main() {
	// 1. Load config
	cfg := config.LoadConfig()

	// 2. Initialize DB
	database := db.InitDB(cfg.Database.Path)

	// 3. Setup Handlers
	h := api.NewHandler(database)

	// 4. Setup Router
	r := api.SetupRouter(h)

	// 5. Start Server
	log.Printf("Starting backend server on port %s", cfg.Server.Port)
	if err := r.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
