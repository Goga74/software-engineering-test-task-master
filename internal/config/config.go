package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	Name    string `yaml:"name"`
	SSLMode string `yaml:"sslmode"`
}

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig `yaml:"database"`
}

// Load reads configuration from config.yaml and applies environment variable overrides
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}

	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_PORT value: %w", err)
		}
		cfg.Database.Port = port
	}

	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}

	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		cfg.Database.SSLMode = sslmode
	}

	return cfg, nil
}

// BuildDSN constructs PostgreSQL connection string from configuration and credentials
func (c *Config) BuildDSN(username, password string) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		username,
		password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetDSN returns the PostgreSQL connection string with backward compatibility
// Priority:
// 1. POSTGRES_DSN environment variable (for backward compatibility)
// 2. Build DSN from config.yaml + environment variables
func GetDSN(configPath string) (string, error) {
	// Check for backward compatibility with existing POSTGRES_DSN
	if dsn := os.Getenv("POSTGRES_DSN"); dsn != "" {
		return dsn, nil
	}

	// Load configuration from file
	cfg, err := Load(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get credentials from environment variables
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	// Validate required credentials
	if username == "" {
		return "", fmt.Errorf("DB_USER environment variable is required when POSTGRES_DSN is not set")
	}
	if password == "" {
		return "", fmt.Errorf("DB_PASSWORD environment variable is required when POSTGRES_DSN is not set")
	}

	// Build and return DSN
	return cfg.BuildDSN(username, password), nil
}
