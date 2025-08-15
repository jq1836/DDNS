package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Server ServerConfig `json:"server"`

	// DDNS specific configuration
	DDNS DDNSConfig `json:"ddns"`

	// HTTP client configuration
	HTTP HTTPConfig `json:"http"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         int      `json:"port"`
	Host         string   `json:"host"`
	ReadTimeout  Duration `json:"read_timeout"`
	WriteTimeout Duration `json:"write_timeout"`
}

// DDNSConfig holds DDNS-related configuration
type DDNSConfig struct {
	Provider       string   `json:"provider"`
	Domain         string   `json:"domain"`
	APIKey         string   `json:"api_key"`
	UpdateInterval Duration `json:"update_interval"`
}

// HTTPConfig holds HTTP client configuration
type HTTPConfig struct {
	Timeout    Duration `json:"timeout"`
	MaxRetries int      `json:"max_retries"`
	RetryDelay Duration `json:"retry_delay"`
	UserAgent  string   `json:"user_agent"`
}

// Duration is a wrapper around time.Duration for JSON unmarshaling
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements json.Unmarshaler for Duration
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	d.Duration = duration
	return nil
}

// MarshalJSON implements json.Marshaler for Duration
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// Load loads configuration from JSON file with fallback to environment variables
func Load() (*Config, error) {
	config := &Config{}

	// Try to load from JSON file first
	if err := loadFromJSON(config); err != nil {
		// If JSON loading fails, fall back to environment variables
		loadFromEnvironment(config)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

// loadFromJSON loads configuration from a JSON file
func loadFromJSON(config *Config) error {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	return nil
}

// loadFromEnvironment loads configuration from environment variables with defaults
func loadFromEnvironment(config *Config) {
	// Load server config
	config.Server = ServerConfig{
		Port:         getEnvAsInt("SERVER_PORT", 8080),
		Host:         getEnv("SERVER_HOST", "localhost"),
		ReadTimeout:  Duration{getEnvAsDuration("SERVER_READ_TIMEOUT", 30*time.Second)},
		WriteTimeout: Duration{getEnvAsDuration("SERVER_WRITE_TIMEOUT", 30*time.Second)},
	}

	// Load DDNS config
	config.DDNS = DDNSConfig{
		Provider:       getEnv("DDNS_PROVIDER", "duckdns"),
		Domain:         getEnv("DDNS_DOMAIN", ""),
		APIKey:         getEnv("DDNS_API_KEY", ""),
		UpdateInterval: Duration{getEnvAsDuration("DDNS_UPDATE_INTERVAL", 5*time.Minute)},
	}

	// Load HTTP config
	config.HTTP = HTTPConfig{
		Timeout:    Duration{getEnvAsDuration("HTTP_TIMEOUT", 30*time.Second)},
		MaxRetries: getEnvAsInt("HTTP_MAX_RETRIES", 3),
		RetryDelay: Duration{getEnvAsDuration("HTTP_RETRY_DELAY", 1*time.Second)},
		UserAgent:  getEnv("HTTP_USER_AGENT", "ddns-client/1.0"),
	}
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}
	return "config.json" // Default config file name
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DDNS.Domain == "" {
		return fmt.Errorf("DDNS domain is required")
	}

	if c.DDNS.APIKey == "" {
		return fmt.Errorf("DDNS API key is required")
	}

	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}

	if c.HTTP.MaxRetries < 0 {
		return fmt.Errorf("HTTP max retries cannot be negative, got %d", c.HTTP.MaxRetries)
	}

	return nil
}

// Helper functions for environment variable parsing

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
