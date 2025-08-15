package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		validate func(*Config) error
	}{
		{
			name: "valid configuration with required fields",
			envVars: map[string]string{
				"DDNS_DOMAIN":  "example.com",
				"DDNS_API_KEY": "test-api-key",
			},
			wantErr: false,
			validate: func(c *Config) error {
				if c.DDNS.Domain != "example.com" {
					t.Errorf("expected domain 'example.com', got '%s'", c.DDNS.Domain)
				}
				if c.DDNS.APIKey != "test-api-key" {
					t.Errorf("expected API key 'test-api-key', got '%s'", c.DDNS.APIKey)
				}
				return nil
			},
		},
		{
			name: "missing required domain",
			envVars: map[string]string{
				"DDNS_API_KEY": "test-api-key",
			},
			wantErr: true,
		},
		{
			name: "missing required API key",
			envVars: map[string]string{
				"DDNS_DOMAIN": "example.com",
			},
			wantErr: true,
		},
		{
			name: "custom values from environment",
			envVars: map[string]string{
				"DDNS_DOMAIN":          "custom.com",
				"DDNS_API_KEY":         "custom-key",
				"DDNS_PROVIDER":        "route53",
				"DDNS_UPDATE_INTERVAL": "10m",
				"SERVER_PORT":          "9090",
				"HTTP_MAX_RETRIES":     "5",
			},
			wantErr: false,
			validate: func(c *Config) error {
				if c.DDNS.Provider != "route53" {
					t.Errorf("expected provider 'route53', got '%s'", c.DDNS.Provider)
				}
				if c.DDNS.UpdateInterval.Duration != 10*time.Minute {
					t.Errorf("expected update interval 10m, got %s", c.DDNS.UpdateInterval.Duration)
				}
				if c.Server.Port != 9090 {
					t.Errorf("expected port 9090, got %d", c.Server.Port)
				}
				if c.HTTP.MaxRetries != 5 {
					t.Errorf("expected max retries 5, got %d", c.HTTP.MaxRetries)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			clearEnv()

			// Temporarily disable JSON config by setting a non-existent path
			os.Setenv("CONFIG_PATH", "non-existent-config.json")
			defer os.Unsetenv("CONFIG_PATH")

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer clearEnv()

			config, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				if err := tt.validate(config); err != nil {
					t.Errorf("validation failed: %v", err)
				}
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				DDNS: DDNSConfig{
					Domain: "example.com",
					APIKey: "test-key",
				},
				Server: ServerConfig{
					Port: 8080,
				},
				HTTP: HTTPConfig{
					MaxRetries: 3,
				},
			},
			wantErr: false,
		},
		{
			name: "missing domain",
			config: &Config{
				DDNS: DDNSConfig{
					APIKey: "test-key",
				},
				Server: ServerConfig{
					Port: 8080,
				},
				HTTP: HTTPConfig{
					MaxRetries: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				DDNS: DDNSConfig{
					Domain: "example.com",
					APIKey: "test-key",
				},
				Server: ServerConfig{
					Port: 99999,
				},
				HTTP: HTTPConfig{
					MaxRetries: 3,
				},
			},
			wantErr: true,
		},
		{
			name: "negative retries",
			config: &Config{
				DDNS: DDNSConfig{
					Domain: "example.com",
					APIKey: "test-key",
				},
				Server: ServerConfig{
					Port: 8080,
				},
				HTTP: HTTPConfig{
					MaxRetries: -1,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to clear environment variables
func clearEnv() {
	envVars := []string{
		"SERVER_PORT", "SERVER_HOST", "SERVER_READ_TIMEOUT", "SERVER_WRITE_TIMEOUT",
		"DDNS_PROVIDER", "DDNS_DOMAIN", "DDNS_API_KEY", "DDNS_UPDATE_INTERVAL",
		"HTTP_TIMEOUT", "HTTP_MAX_RETRIES", "HTTP_RETRY_DELAY", "HTTP_USER_AGENT",
		"CONFIG_PATH",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}
