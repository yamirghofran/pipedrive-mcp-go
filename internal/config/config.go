// Package config handles environment variable parsing and validation
// for the Pipedrive MCP server.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// DefaultBookingFieldKey is the default Pipedrive custom field key for booking details.
const DefaultBookingFieldKey = "8f4b27fbd9dfc70d2296f23ce76987051ad7324e"

// Config holds all configuration for the MCP server.
type Config struct {
	// Required
	PipedriveAPIToken string
	PipedriveDomain   string

	// Transport
	MCPTransport string // "stdio" or "http"
	MCPPort      int
	MCPHost      string

	// JWT Authentication (HTTP transport only)
	MCPJWTSecret    string
	MCPJWTAlgorithm string
	MCPJWTAudience  string
	MCPJWTIssuer    string

	// Rate Limiting
	RateLimitMinInterval   time.Duration
	RateLimitMaxConcurrent int

	// Session Management (HTTP transport)
	SessionMaxAge          time.Duration
	SessionCleanupInterval time.Duration
	SessionMaxCount        int

	// Custom Field
	BookingFieldKey string

	// TLS (optional)
	TLSCert string
	TLSKey  string
}

// Load reads configuration from environment variables and applies defaults.
func Load() (*Config, error) {
	cfg := &Config{
		// Required
		PipedriveAPIToken: os.Getenv("PIPEDRIVE_API_TOKEN"),
		PipedriveDomain:   os.Getenv("PIPEDRIVE_DOMAIN"),

		// Transport defaults
		MCPTransport: getEnvOrDefault("MCP_TRANSPORT", "stdio"),
		MCPPort:      getEnvIntOrDefault("MCP_PORT", 3000),
		MCPHost:      getEnvOrDefault("MCP_HOST", "localhost"),

		// JWT defaults
		MCPJWTSecret:    os.Getenv("MCP_JWT_SECRET"),
		MCPJWTAlgorithm: getEnvOrDefault("MCP_JWT_ALGORITHM", "HS256"),
		MCPJWTAudience:  os.Getenv("MCP_JWT_AUDIENCE"),
		MCPJWTIssuer:    os.Getenv("MCP_JWT_ISSUER"),

		// Rate limiting defaults
		RateLimitMinInterval:   getEnvDurationOrDefault("PIPEDRIVE_RATE_LIMIT_MIN_INTERVAL", 250*time.Millisecond),
		RateLimitMaxConcurrent: getEnvIntOrDefault("PIPEDRIVE_RATE_LIMIT_MAX_CONCURRENT", 2),

		// Session defaults
		SessionMaxAge:          getEnvDurationOrDefault("MCP_SESSION_MAX_AGE", 1*time.Hour),
		SessionCleanupInterval: getEnvDurationOrDefault("MCP_SESSION_CLEANUP_INTERVAL", 5*time.Minute),
		SessionMaxCount:        getEnvIntOrDefault("MCP_SESSION_MAX_COUNT", 100),

		// Custom field
		BookingFieldKey: getEnvOrDefault("PIPEDRIVE_BOOKING_FIELD_KEY", DefaultBookingFieldKey),

		// TLS
		TLSCert: os.Getenv("MCP_TLS_CERT"),
		TLSKey:  os.Getenv("MCP_TLS_KEY"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.PipedriveAPIToken == "" {
		return fmt.Errorf("PIPEDRIVE_API_TOKEN environment variable is required")
	}
	if c.PipedriveDomain == "" {
		return fmt.Errorf("PIPEDRIVE_DOMAIN environment variable is required")
	}

	switch c.MCPTransport {
	case "stdio", "http":
		// valid
	default:
		return fmt.Errorf("MCP_TRANSPORT must be 'stdio' or 'http', got %q", c.MCPTransport)
	}

	if c.MCPJWTAlgorithm != "" {
		switch c.MCPJWTAlgorithm {
		case "HS256", "HS384", "HS512", "RS256", "RS384", "RS512":
			// valid
		default:
			return fmt.Errorf("MCP_JWT_ALGORITHM must be one of HS256, HS384, HS512, RS256, RS384, RS512, got %q", c.MCPJWTAlgorithm)
		}
	}

	if c.MCPPort < 1 || c.MCPPort > 65535 {
		return fmt.Errorf("MCP_PORT must be between 1 and 65535, got %d", c.MCPPort)
	}

	if c.SessionMaxCount < 1 {
		return fmt.Errorf("MCP_SESSION_MAX_COUNT must be at least 1, got %d", c.SessionMaxCount)
	}

	return nil
}

// IsHTTPTransport returns true if the server should use HTTP transport.
func (c *Config) IsHTTPTransport() bool {
	return c.MCPTransport == "http"
}

// HasJWTAuth returns true if JWT authentication is configured.
func (c *Config) HasJWTAuth() bool {
	return c.MCPJWTSecret != ""
}

// HasTLS returns true if TLS certificates are configured.
func (c *Config) HasTLS() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}

// ListenAddr returns the host:port string for the HTTP server.
func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.MCPHost, c.MCPPort)
}

// --- helper functions ---

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvIntOrDefault(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func getEnvDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultVal
	}
	return d
}
