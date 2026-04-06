package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yamirghofran/pipedrive-mcp-go/internal/middleware"
)

// SessionInfo tracks an HTTP session with its creation time.
type SessionInfo struct {
	CreatedAt time.Time
}

// HTTPTransport manages the HTTP transport for the MCP server.
type HTTPTransport struct {
	server                 *mcp.Server
	addr                   string
	sessions               sync.Map
	sessionMaxAge          time.Duration
	sessionCleanupInterval time.Duration
	sessionMaxCount        int
	jwtSecret              string
	jwtAlgorithm           string
	jwtAudience            string
	jwtIssuer              string
	corsOrigins            []string
	tlsCert                string
	tlsKey                 string
}

// NewHTTPTransport creates a new HTTP transport manager.
func NewHTTPTransport(
	server *mcp.Server,
	addr string,
	sessionMaxAge, sessionCleanupInterval time.Duration,
	sessionMaxCount int,
	jwtSecret, jwtAlgorithm, jwtAudience, jwtIssuer string,
	corsOrigins []string,
	tlsCert, tlsKey string,
) *HTTPTransport {
	return &HTTPTransport{
		server:                 server,
		addr:                   addr,
		sessionMaxAge:          sessionMaxAge,
		sessionCleanupInterval: sessionCleanupInterval,
		sessionMaxCount:        sessionMaxCount,
		jwtSecret:              jwtSecret,
		jwtAlgorithm:           jwtAlgorithm,
		jwtAudience:            jwtAudience,
		jwtIssuer:              jwtIssuer,
		corsOrigins:            corsOrigins,
		tlsCert:                tlsCert,
		tlsKey:                 tlsKey,
	}
}

// Run starts the HTTP server.
func (t *HTTPTransport) Run(ctx context.Context) error {
	// Create the streamable HTTP handler
	mcpHandler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return t.server
	}, nil)

	// Build the HTTP handler chain
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "ok",
			"transport": "http",
		})
	})

	// MCP endpoint handler
	mcpEndpoint := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track sessions
		sessionID := r.Header.Get("Mcp-Session-Id")
		if sessionID != "" {
			t.sessions.Store(sessionID, &SessionInfo{CreatedAt: time.Now()})
		}

		mcpHandler.ServeHTTP(w, r)
	})

	// Apply middleware chain
	var h http.Handler = mcpEndpoint

	// JWT auth middleware (if configured)
	if t.jwtSecret != "" {
		authMW, err := middleware.NewJWTAuth(t.jwtSecret, t.jwtAlgorithm, t.jwtAudience, t.jwtIssuer)
		if err != nil {
			return fmt.Errorf("failed to create JWT middleware: %w", err)
		}
		h = authMW.Middleware(h)
	}

	// CORS middleware
	h = middleware.CORS(t.corsOrigins)(h)

	// Request body size limit
	h = middleware.MaxBodySize(1 << 20)(h) // 1MB

	mux.Handle("/mcp", h)

	// Start session cleanup goroutine
	go t.cleanupSessions(ctx)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:    t.addr,
		Handler: mux,
	}

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		slog.Info("Starting MCP server with HTTP transport", "addr", t.addr)
		if t.tlsCert != "" && t.tlsKey != "" {
			errCh <- httpServer.ListenAndServeTLS(t.tlsCert, t.tlsKey)
		} else {
			errCh <- httpServer.ListenAndServe()
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		slog.Info("Shutting down HTTP server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return httpServer.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != http.ErrServerClosed {
			return err
		}
		return nil
	}
}

// cleanupSessions periodically removes expired sessions.
func (t *HTTPTransport) cleanupSessions(ctx context.Context) {
	ticker := time.NewTicker(t.sessionCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var count int
			now := time.Now()
			t.sessions.Range(func(key, value interface{}) bool {
				info, ok := value.(*SessionInfo)
				if !ok {
					t.sessions.Delete(key)
					return true
				}
				if now.Sub(info.CreatedAt) > t.sessionMaxAge {
					t.sessions.Delete(key)
					count++
				}
				return true
			})
			if count > 0 {
				slog.Info("Cleaned up expired sessions", "count", count)
			}

			// Log current session count
			var total int
			t.sessions.Range(func(_, _ interface{}) bool {
				total++
				return true
			})
			if total > 0 {
				slog.Debug("Active sessions", "count", total, "max", t.sessionMaxCount)
			}
		}
	}
}

// DefaultCORSHosts returns the default CORS allowed origins.
func DefaultCORSHosts() []string {
	return []string{
		"http://localhost:*",
		"http://127.0.0.1:*",
	}
}

// ParseCORSHosts parses a comma-separated list of CORS hosts.
func ParseCORSHosts(hosts string) []string {
	if hosts == "" {
		return DefaultCORSHosts()
	}
	parts := strings.Split(hosts, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return DefaultCORSHosts()
	}
	return result
}
