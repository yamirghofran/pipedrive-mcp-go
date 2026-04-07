// Package server creates and configures the MCP server with all tools and prompts.
package server

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yamirghofran/pipedrive-mcp-go/internal/pipedrive"
)

const (
	ServerName    = "pipedrive-mcp-server"
	ServerVersion = "3.0.0"
)

// NewServer creates a new MCP server with all Pipedrive tools and prompts registered.
func NewServer(client *pipedrive.Client) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    ServerName,
		Version: ServerVersion,
	}, nil)

	registerTools(server, client)
	registerPrompts(server)

	return server
}
