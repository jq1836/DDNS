package providers

import (
	"context"
	"fmt"
	"time"

	"ddns/ddns"
)

// MockProvider is a simple mock implementation for testing
type MockProvider struct {
	name           string
	records        map[string]string // domain -> IP mapping
	shouldFail     bool
	validateResult error
}

// NewMockProvider creates a new mock DDNS provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:    name,
		records: make(map[string]string),
	}
}

// WithFailure configures the mock to fail operations
func (m *MockProvider) WithFailure(shouldFail bool) *MockProvider {
	m.shouldFail = shouldFail
	return m
}

// WithValidationError configures the mock to fail validation
func (m *MockProvider) WithValidationError(err error) *MockProvider {
	m.validateResult = err
	return m
}

// UpdateRecord updates a DNS record (mock implementation)
func (m *MockProvider) UpdateRecord(ctx context.Context, req ddns.UpdateRequest) (*ddns.UpdateResponse, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock provider configured to fail")
	}

	key := fmt.Sprintf("%s:%s", req.Domain, req.RecordType)
	m.records[key] = req.Value

	return &ddns.UpdateResponse{
		Success:   true,
		Message:   fmt.Sprintf("Mock update successful for %s", req.Domain),
		RecordID:  fmt.Sprintf("mock-record-%d", time.Now().Unix()),
		UpdatedAt: time.Now(),
	}, nil
}

// GetCurrentRecord retrieves the current DNS record value (mock implementation)
func (m *MockProvider) GetCurrentRecord(ctx context.Context, domain, recordType string) (string, error) {
	if m.shouldFail {
		return "", fmt.Errorf("mock provider configured to fail")
	}

	key := fmt.Sprintf("%s:%s", domain, recordType)
	if value, exists := m.records[key]; exists {
		return value, nil
	}

	return "", fmt.Errorf("record not found")
}

// ValidateCredentials checks if the provider credentials are valid (mock implementation)
func (m *MockProvider) ValidateCredentials(ctx context.Context) error {
	if m.validateResult != nil {
		return m.validateResult
	}

	if m.shouldFail {
		return fmt.Errorf("mock validation failed")
	}

	return nil
}

// GetProviderName returns the name of the DDNS provider
func (m *MockProvider) GetProviderName() string {
	return fmt.Sprintf("mock-%s", m.name)
}

// SetRecord manually sets a record (for testing)
func (m *MockProvider) SetRecord(domain, recordType, value string) {
	key := fmt.Sprintf("%s:%s", domain, recordType)
	m.records[key] = value
}

// GetRecords returns all stored records (for testing)
func (m *MockProvider) GetRecords() map[string]string {
	return m.records
}
