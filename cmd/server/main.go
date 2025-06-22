package main

import (
	"log"
	"super-payment/internal/api"
	"super-payment/internal/config"
	"super-payment/internal/repository"
	"super-payment/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize repository
	repo, err := repository.NewMySQLRepository(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing repository: %v", err)
		}
	}()

	// Initialize service
	svc := service.NewInvoiceService(repo)

	// Initialize HTTP handler
	handler := api.NewHandler(svc, cfg)

	// Setup routes
	router := handler.SetupRoutes()

	// Start server
	serverAddr := cfg.GetServerAddress()
	log.Printf("Starting server on %s", serverAddr)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
