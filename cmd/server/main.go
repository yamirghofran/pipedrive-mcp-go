// pipedrive-mcp-server is a Go-based MCP server for Pipedrive CRM.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yamirghofran/pipedrive-mcp-go/internal/config"
	"github.com/yamirghofran/pipedrive-mcp-go/internal/pipedrive"
	"github.com/yamirghofran/pipedrive-mcp-go/internal/server"
	"github.com/yamirghofran/pipedrive-mcp-go/internal/transport"
)

func main() {
	// Configure structured JSON logging to stderr (never stdout - reserved for MCP protocol in stdio mode)
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Print startup banner
	fmt.Fprintf(os.Stderr, "pipedrive-mcp-server %s\n", server.ServerVersion)

	// Load configuration from environment variables
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Configuration error", "error", err)
		os.Exit(1)
	}

	slog.Info("Configuration loaded",
		"transport", cfg.MCPTransport,
		"domain", cfg.PipedriveDomain,
	)

	// Create Pipedrive API client
	client := pipedrive.NewClient(
		cfg.PipedriveDomain,
		cfg.PipedriveAPIToken,
		cfg.RateLimitMinInterval,
		cfg.RateLimitMaxConcurrent,
		cfg.BookingFieldKey,
	)

	// Create MCP server with all tools and prompts
	mcpServer := server.NewServer(client)

	slog.Info("MCP server created",
		"name", server.ServerName,
		"version", server.ServerVersion,
	)

	// Setup context with signal handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start server with appropriate transport
	if cfg.IsHTTPTransport() {
		runHTTP(ctx, cfg, mcpServer)
	} else {
		runStdio(ctx, mcpServer)
	}
}

func runStdio(ctx context.Context, mcpServer *mcp.Server) {
	slog.Info("Starting with stdio transport (reading from stdin, writing to stdout)")

	if err := transport.RunStdio(ctx, mcpServer); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func runHTTP(ctx context.Context, cfg *config.Config, mcpServer *mcp.Server) {
	corsHosts := transport.DefaultCORSHosts()

	httpTransport := transport.NewHTTPTransport(
		mcpServer,
		cfg.ListenAddr(),
		cfg.SessionMaxAge,
		cfg.SessionCleanupInterval,
		cfg.SessionMaxCount,
		cfg.MCPJWTSecret,
		cfg.MCPJWTAlgorithm,
		cfg.MCPJWTAudience,
		cfg.MCPJWTIssuer,
		corsHosts,
		cfg.TLSCert,
		cfg.TLSKey,
	)

	slog.Info("Starting HTTP transport",
		"addr", cfg.ListenAddr(),
		"jwt_auth", cfg.HasJWTAuth(),
		"tls", cfg.HasTLS(),
	)

	if err := httpTransport.Run(ctx); err != nil {
		slog.Error("HTTP server failed", "error", err)
		os.Exit(1)
	}
}
