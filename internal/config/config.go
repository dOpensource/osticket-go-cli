package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var cfg *viper.Viper

// Environment variable names
const (
	EnvBaseURL = "OSTICKET_BASE_URL"
	EnvAPIKey  = "OSTICKET_API_KEY"
)

func init() {
	cfg = viper.New()
	cfg.SetConfigName("config")
	cfg.SetConfigType("yaml")

	// Config directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	configDir := filepath.Join(homeDir, ".osticket-cli")
	cfg.AddConfigPath(configDir)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not create config directory: %v\n", err)
	}

	// Set defaults
	cfg.SetDefault("base_url", "")
	cfg.SetDefault("api_key", "")

	// Bind environment variables
	cfg.BindEnv("base_url", EnvBaseURL)
	cfg.BindEnv("api_key", EnvAPIKey)

	// Read config file if it exists
	if err := cfg.ReadInConfig(); err != nil {
		// Config file not found is okay
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: error reading config: %v\n", err)
		}
	}
}

// Get returns a config value
func Get(key string) string {
	return cfg.GetString(key)
}

// Set sets a config value and saves to file
func Set(key, value string) error {
	cfg.Set(key, value)
	return Save()
}

// Save writes the config to file
func Save() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".osticket-cli", "config.yaml")
	return cfg.WriteConfigAs(configPath)
}

// GetBaseURL returns the API base URL (env var takes precedence)
func GetBaseURL() string {
	// Check environment variable first
	if envVal := os.Getenv(EnvBaseURL); envVal != "" {
		return envVal
	}
	return cfg.GetString("base_url")
}

// GetAPIKey returns the API key (env var takes precedence)
func GetAPIKey() string {
	// Check environment variable first
	if envVal := os.Getenv(EnvAPIKey); envVal != "" {
		return envVal
	}
	return cfg.GetString("api_key")
}

// SetBaseURL sets the API base URL
func SetBaseURL(url string) error {
	return Set("base_url", url)
}

// SetAPIKey sets the API key
func SetAPIKey(key string) error {
	return Set("api_key", key)
}

// IsConfigured checks if the CLI is configured
func IsConfigured() bool {
	return GetBaseURL() != "" && GetAPIKey() != ""
}

// Clear clears all configuration
func Clear() error {
	cfg.Set("base_url", "")
	cfg.Set("api_key", "")
	return Save()
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".osticket-cli", "config.yaml")
}

// GetConfigSource returns where each config value is coming from
func GetConfigSource() (baseURLSource, apiKeySource string) {
	if os.Getenv(EnvBaseURL) != "" {
		baseURLSource = "env:" + EnvBaseURL
	} else if cfg.GetString("base_url") != "" {
		baseURLSource = "config"
	} else {
		baseURLSource = "not set"
	}

	if os.Getenv(EnvAPIKey) != "" {
		apiKeySource = "env:" + EnvAPIKey
	} else if cfg.GetString("api_key") != "" {
		apiKeySource = "config"
	} else {
		apiKeySource = "not set"
	}

	return
}
