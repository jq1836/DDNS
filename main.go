package main

import (
	"context"
	"ddns/config"
	"ddns/ddns"
	"ddns/providers"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load and validate configuration
	cfg := loadAndValidateConfig()

	// Setup DDNS service
	service := setupDDNSService(cfg)

	// Run the DDNS client
	runDDNSClient(service, cfg.DDNS.UpdateInterval.Duration)
}

func loadAndValidateConfig() *config.Config {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Printf("Starting DDNS client for domain: %s", cfg.DDNS.Domain)
	log.Printf("Using provider: %s", cfg.DDNS.Provider)
	log.Printf("Update interval: %s", cfg.DDNS.UpdateInterval.Duration)

	return cfg
}

func setupDDNSService(cfg *config.Config) *ddns.Service {
	// Create provider factory
	factory := providers.NewFactory()

	// Create DDNS config
	ddnsConfig := ddns.Config{
		Provider:   cfg.DDNS.Provider,
		APIKey:     cfg.DDNS.APIKey,
		Domain:     cfg.DDNS.Domain,
		TTL:        300, // Default TTL
		RecordType: "A", // Default to A record
	}

	// Create provider
	provider, err := factory.CreateProvider(ddnsConfig)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	// Validate provider credentials
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := provider.ValidateCredentials(ctx); err != nil {
		log.Fatalf("Provider credential validation failed: %v", err)
	}

	log.Printf("Provider credentials validated successfully")

	// Create and return DDNS service
	return ddns.NewService(provider, ddnsConfig)
}

func setupGracefulShutdown() (context.Context, context.CancelFunc) {
	mainCtx, mainCancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping...")
		mainCancel()
	}()

	return mainCtx, mainCancel
}

func performDDNSUpdate(ctx context.Context, service *ddns.Service) {
	updateCtx, updateCancel := context.WithTimeout(ctx, 2*time.Minute)
	defer updateCancel()

	log.Println("Checking for IP changes...")
	response, err := service.UpdateIP(updateCtx)
	if err != nil {
		log.Printf("Failed to update IP: %v", err)
		return
	}

	if response.Success {
		log.Printf("DNS update successful: %s", response.Message)
	} else {
		log.Printf("DNS update failed: %s", response.Message)
	}

	if response.RecordID != "" {
		log.Printf("Record ID: %s", response.RecordID)
	}
}

func runDDNSClient(service *ddns.Service, updateInterval time.Duration) {
	// Setup graceful shutdown
	mainCtx, mainCancel := setupGracefulShutdown()
	defer mainCancel()

	// Create ticker for periodic updates
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	// Perform initial update
	log.Println("Performing initial IP update...")
	performDDNSUpdate(mainCtx, service)

	// Start the update loop
	for {
		select {
		case <-mainCtx.Done():
			log.Println("DDNS client stopped")
			return
		case <-ticker.C:
			performDDNSUpdate(mainCtx, service)
		}
	}
}
