package transport

import (
	"context"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunStdio starts the MCP server on stdio transport.
func RunStdio(ctx context.Context, server *mcp.Server) error {
	slog.Info("Starting MCP server with stdio transport")
	return server.Run(ctx, &mcp.StdioTransport{})
}
