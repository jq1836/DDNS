package ddns

import (
	"context"
	"testing"
	"time"
)

// mockProvider for testing
type mockProvider struct {
	name           string
	records        map[string]string
	shouldFail     bool
	validateResult error
}

// mockIPDetector for testing
type mockIPDetector struct {
	ip         string
	shouldFail bool
}

func (m *mockIPDetector) GetPublicIP(ctx context.Context) (string, error) {
	if m.shouldFail {
		return "", &mockError{"IP detection failed"}
	}
	return m.ip, nil
}

func newMockProvider(name string) *mockProvider {
	return &mockProvider{
		name:    name,
		records: make(map[string]string),
	}
}

func (m *mockProvider) UpdateRecord(ctx context.Context, req UpdateRequest) (*UpdateResponse, error) {
	if m.shouldFail {
		return nil, &mockError{"update failed"}
	}

	key := req.Domain + ":" + req.RecordType
	m.records[key] = req.Value

	return &UpdateResponse{
		Success:   true,
		Message:   "Updated successfully",
		RecordID:  "mock-123",
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockProvider) GetCurrentRecord(ctx context.Context, domain, recordType string) (string, error) {
	if m.shouldFail {
		return "", &mockError{"get record failed"}
	}

	key := domain + ":" + recordType
	if value, exists := m.records[key]; exists {
		return value, nil
	}
	return "", &mockError{"record not found"}
}

func (m *mockProvider) ValidateCredentials(ctx context.Context) error {
	return m.validateResult
}

func (m *mockProvider) GetProviderName() string {
	return m.name
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestServiceUpdateIP(t *testing.T) {
	provider := newMockProvider("test")
	config := Config{
		Domain:     "example.com",
		RecordType: "A",
		TTL:        300,
	}

	// Mock IP detector that returns a known IP
	ipDetector := &mockIPDetector{ip: "203.0.113.1"}
	service := NewServiceWithIPDetector(provider, config, ipDetector)

	// Test successful update
	resp, err := service.UpdateIP(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Error("Expected successful update")
	}

	if resp.Message == "" {
		t.Error("Expected non-empty message")
	}

	// Verify the record was updated with the mocked IP
	key := config.Domain + ":" + config.RecordType
	if provider.records[key] != "203.0.113.1" {
		t.Errorf("Expected record to be updated with IP 203.0.113.1, got %s", provider.records[key])
	}
}

func TestServiceUpdateIPNoChangeNeeded(t *testing.T) {
	provider := newMockProvider("test")
	config := Config{
		Domain:     "example.com",
		RecordType: "A",
		TTL:        300,
	}

	// Pre-populate with current IP that matches what the IP detector will return
	currentIP := "203.0.113.1"
	provider.records["example.com:A"] = currentIP

	// Mock IP detector that returns the same IP as the existing record
	ipDetector := &mockIPDetector{ip: currentIP}
	service := NewServiceWithIPDetector(provider, config, ipDetector)

	resp, err := service.UpdateIP(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !resp.Success {
		t.Error("Expected successful response")
	}

	// Should indicate no change was needed
	if resp.Message != "Record already up to date" {
		t.Errorf("Expected 'Record already up to date' message, got %s", resp.Message)
	}
}

func TestServiceUpdateIPDetectionFails(t *testing.T) {
	provider := newMockProvider("test")
	config := Config{
		Domain:     "example.com",
		RecordType: "A",
		TTL:        300,
	}

	// Mock IP detector that fails
	ipDetector := &mockIPDetector{shouldFail: true}
	service := NewServiceWithIPDetector(provider, config, ipDetector)

	resp, err := service.UpdateIP(context.Background())
	if err == nil {
		t.Fatal("Expected error due to IP detection failure")
	}

	if resp != nil {
		t.Error("Expected nil response when IP detection fails")
	}
}

func TestServiceValidate(t *testing.T) {
	provider := newMockProvider("test")
	config := Config{}

	service := NewService(provider, config)

	// Test successful validation
	err := service.Validate(context.Background())
	if err != nil {
		t.Errorf("Expected no validation error, got %v", err)
	}

	// Test failed validation
	provider.validateResult = &mockError{"invalid credentials"}
	err = service.Validate(context.Background())
	if err == nil {
		t.Error("Expected validation error")
	}
}

func TestServiceGetProvider(t *testing.T) {
	provider := newMockProvider("test")
	config := Config{}

	service := NewService(provider, config)

	if service.GetProvider() != provider {
		t.Error("GetProvider should return the configured provider")
	}

	if service.GetProvider().GetProviderName() != "test" {
		t.Error("Provider name should match")
	}
}

func TestUpdateRequest(t *testing.T) {
	req := UpdateRequest{
		Domain:     "test.duckdns.org",
		RecordType: "A",
		Value:      "192.168.1.100",
		TTL:        600,
	}

	if req.Domain != "test.duckdns.org" {
		t.Error("Domain not set correctly")
	}

	if req.RecordType != "A" {
		t.Error("RecordType not set correctly")
	}

	if req.Value != "192.168.1.100" {
		t.Error("Value not set correctly")
	}

	if req.TTL != 600 {
		t.Error("TTL not set correctly")
	}
}

func TestUpdateResponse(t *testing.T) {
	now := time.Now()
	resp := UpdateResponse{
		Success:   true,
		Message:   "Test message",
		RecordID:  "test-123",
		UpdatedAt: now,
	}

	if !resp.Success {
		t.Error("Success not set correctly")
	}

	if resp.Message != "Test message" {
		t.Error("Message not set correctly")
	}

	if resp.RecordID != "test-123" {
		t.Error("RecordID not set correctly")
	}

	if !resp.UpdatedAt.Equal(now) {
		t.Error("UpdatedAt not set correctly")
	}
}

func TestConfig(t *testing.T) {
	config := Config{
		Provider:       "duckdns",
		APIKey:         "test-key",
		Domain:         "example.duckdns.org",
		TTL:            300,
		RecordType:     "A",
		UpdateInterval: 5 * time.Minute,
	}

	if config.Provider != "duckdns" {
		t.Error("Provider not set correctly")
	}

	if config.APIKey != "test-key" {
		t.Error("APIKey not set correctly")
	}

	if config.Domain != "example.duckdns.org" {
		t.Error("Domain not set correctly")
	}

	if config.TTL != 300 {
		t.Error("TTL not set correctly")
	}

	if config.UpdateInterval != 5*time.Minute {
		t.Error("UpdateInterval not set correctly")
	}
}
