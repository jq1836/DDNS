package providers

import (
	"fmt"

	"github.com/jq1836/DDNS/ddns"
)

// Factory creates DDNS providers based on configuration
type Factory struct{}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProvider creates a DDNS provider based on the configuration
func (f *Factory) CreateProvider(config ddns.Config) (ddns.Provider, error) {
	switch config.Provider {
	case "duckdns":
		if config.APIKey == "" {
			return nil, fmt.Errorf("duckdns provider requires API key (token)")
		}

		duckConfig := DuckDNSConfig{
			Token: config.APIKey,
		}

		return NewDuckDNSProvider(duckConfig), nil

	case "mock":
		return NewMockProvider("test"), nil

	default:
		return nil, fmt.Errorf("unsupported DDNS provider: %s", config.Provider)
	}
}

// GetSupportedProviders returns a list of supported provider names
func (f *Factory) GetSupportedProviders() []string {
	return []string{
		"duckdns",
		"mock",
	}
}

// ValidateProviderConfig validates the configuration for a specific provider
func (f *Factory) ValidateProviderConfig(config ddns.Config) error {
	switch config.Provider {
	case "duckdns":
		if config.APIKey == "" {
			return fmt.Errorf("duckdns provider requires API key (token)")
		}
		return nil

	case "mock":
		// Mock provider doesn't require any specific configuration
		return nil

	default:
		return fmt.Errorf("unsupported DDNS provider: %s", config.Provider)
	}
}
