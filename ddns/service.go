package ddns

import (
	"context"
	"time"
)

// UpdateRequest represents a DDNS update request
type UpdateRequest struct {
	Domain     string
	RecordType string // A, AAAA, CNAME, etc.
	Value      string // IP address or target value
	TTL        int    // Time to live in seconds
}

// UpdateResponse represents the response from a DDNS update
type UpdateResponse struct {
	Success   bool
	Message   string
	RecordID  string // Provider-specific record identifier
	UpdatedAt time.Time
}

// Provider defines the interface that all DDNS providers must implement
type Provider interface {
	// UpdateRecord updates a DNS record for the given domain
	UpdateRecord(ctx context.Context, req UpdateRequest) (*UpdateResponse, error)

	// GetCurrentRecord retrieves the current DNS record value
	GetCurrentRecord(ctx context.Context, domain, recordType string) (string, error)

	// ValidateCredentials checks if the provider credentials are valid
	ValidateCredentials(ctx context.Context) error

	// GetProviderName returns the name of the DDNS provider
	GetProviderName() string
}

// Config holds configuration for DDNS providers
type Config struct {
	Provider string
	APIKey   string // This will be the token for DuckDNS
	Domain   string
	TTL      int

	// Additional settings
	RecordType     string
	UpdateInterval time.Duration
}

// Service manages DDNS updates using the configured provider
type Service struct {
	provider Provider
	config   Config
}

// NewService creates a new DDNS service with the specified provider
func NewService(provider Provider, config Config) *Service {
	return &Service{
		provider: provider,
		config:   config,
	}
}

// UpdateIP updates the DNS record with the current public IP
func (s *Service) UpdateIP(ctx context.Context) (*UpdateResponse, error) {
	// Get current public IP
	currentIP, err := s.getCurrentPublicIP(ctx)
	if err != nil {
		return nil, err
	}

	// Check if update is needed
	existingRecord, err := s.provider.GetCurrentRecord(ctx, s.config.Domain, s.config.RecordType)
	if err == nil && existingRecord == currentIP {
		// No update needed
		return &UpdateResponse{
			Success:   true,
			Message:   "Record already up to date",
			UpdatedAt: time.Now(),
		}, nil
	}

	// Update the record
	req := UpdateRequest{
		Domain:     s.config.Domain,
		RecordType: s.config.RecordType,
		Value:      currentIP,
		TTL:        s.config.TTL,
	}

	return s.provider.UpdateRecord(ctx, req)
}

// GetCurrentPublicIP retrieves the current public IP address
func (s *Service) getCurrentPublicIP(ctx context.Context) (string, error) {
	// This could be enhanced to support multiple IP detection services
	// For now, we'll use a simple approach
	return getCurrentPublicIPFromService(ctx)
}

// Validate checks if the service configuration and credentials are valid
func (s *Service) Validate(ctx context.Context) error {
	return s.provider.ValidateCredentials(ctx)
}

// GetProvider returns the underlying provider
func (s *Service) GetProvider() Provider {
	return s.provider
}

// getCurrentPublicIPFromService gets the public IP from an external service
func getCurrentPublicIPFromService(ctx context.Context) (string, error) {
	// Simple implementation - in practice you might want to try multiple services
	// and use the executor for retry logic
	return getIPFromHTTPBin(ctx)
}
