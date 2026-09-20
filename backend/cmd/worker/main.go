package main

import (
	"context"
	"log"
	"os"

	"studentos/backend/internal/config"
	"studentos/backend/internal/database"
	"studentos/backend/internal/ingest"
)

func main() {
	log.Println("[INFO] Starting StudentOS Ingestion Worker...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[FATAL] Invalid configuration: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), ingest.Timeout)
	defer cancel()

	res, err := ingest.Run(ctx, db)
	if err != nil {
		log.Fatalf("[FATAL] Ingestion: %v", err)
	}
	if res.TotalOutage() {
		os.Exit(1) // surface a total outage as a failed cron run
	}
}
