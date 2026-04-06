package server

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Prompt definitions matching the spec exactly.
var promptDefs = []struct {
	name        string
	description string
	message     string
}{
	{
		name:        "list-all-deals",
		description: "List all deals in Pipedrive",
		message:     "Please list all deals in my Pipedrive account, showing their title, value, status, and stage.",
	},
	{
		name:        "list-all-persons",
		description: "List all persons in Pipedrive",
		message:     "Please list all persons in my Pipedrive account, showing their name, email, phone, and organization.",
	},
	{
		name:        "list-all-pipelines",
		description: "List all pipelines in Pipedrive",
		message:     "Please list all pipelines in my Pipedrive account, showing their name and stages.",
	},
	{
		name:        "analyze-deals",
		description: "Analyze deals by stage",
		message:     "Please analyze the deals in my Pipedrive account, grouping them by stage and providing total value for each stage.",
	},
	{
		name:        "analyze-contacts",
		description: "Analyze contacts by organization",
		message:     "Please analyze the persons in my Pipedrive account, grouping them by organization and providing a count for each organization.",
	},
	{
		name:        "analyze-leads",
		description: "Analyze leads by status",
		message:     "Please search for all leads in my Pipedrive account and group them by status.",
	},
	{
		name:        "compare-pipelines",
		description: "Compare different pipelines and their stages",
		message:     "Please list all pipelines in my Pipedrive account and compare them by showing the stages in each pipeline.",
	},
	{
		name:        "find-high-value-deals",
		description: "Find high-value deals",
		message:     "Please identify the highest value deals in my Pipedrive account and provide information about which stage they're in and which person or organization they're associated with.",
	},
}

// registerPrompts registers all static MCP prompts.
func registerPrompts(s *mcp.Server) {
	for _, def := range promptDefs {
		prompt := &mcp.Prompt{
			Name:        def.name,
			Description: def.description,
		}

		// Capture the message for the closure
		msg := def.message
		handler := func(_ context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{
				Description: def.description,
				Messages: []*mcp.PromptMessage{
					{
						Role:    "user",
						Content: &mcp.TextContent{Text: msg},
					},
				},
			}, nil
		}

		s.AddPrompt(prompt, handler)
	}
}
