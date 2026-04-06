package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/yamirghofran/pipedrive-mcp-go/internal/pipedrive"
)

// toolError creates an MCP error result with a sanitized message.
func toolError(format string, args ...interface{}) (*mcp.CallToolResult, any, error) {
	msg := fmt.Sprintf(format, args...)
	slog.Error("tool error", "message", msg)
	result := &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}
	result.SetError(fmt.Errorf("%s", msg))
	return result, nil, nil
}

// toolSuccess creates a successful MCP result with JSON data.
func toolSuccess(data interface{}) (*mcp.CallToolResult, any, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return toolError("failed to serialize response: %v", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil, nil
}

// toolSuccessRaw creates a successful MCP result from raw JSON bytes.
func toolSuccessRaw(data json.RawMessage) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// --- Input types for tools ---

type emptyInput struct{}

type dealIDInput struct {
	DealID int `json:"dealId" jsonschema:"Pipedrive deal ID,required"`
}

type personIDInput struct {
	PersonID int `json:"personId" jsonschema:"Pipedrive person ID,required"`
}

type orgIDInput struct {
	OrganizationID int `json:"organizationId" jsonschema:"Pipedrive organization ID,required"`
}

type searchInput struct {
	Term string `json:"term" jsonschema:"Search term,required"`
}

type dealNotesInput struct {
	DealID int  `json:"dealId" jsonschema:"Pipedrive deal ID,required"`
	Limit  *int `json:"limit,omitempty" jsonschema:"Maximum notes to return. Default: 20"`
}

type searchAllInput struct {
	Term      string  `json:"term" jsonschema:"Search term,required"`
	ItemTypes *string `json:"itemTypes,omitempty" jsonschema:"Comma-separated item types: deal, person, organization, product, file, activity, lead"`
}

// registerTools registers all 15 MCP tools.
func registerTools(s *mcp.Server, client *pipedrive.Client) {
	// Tool 1: get-users
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-users",
		Description: "Get all users/owners from Pipedrive to identify owner IDs for filtering deals",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		users, err := client.GetUsers(ctx)
		if err != nil {
			return toolError("Failed to fetch users: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccess(map[string]interface{}{
			"summary": fmt.Sprintf("Found %d users in your Pipedrive account", len(users)),
			"users":   users,
		})
	})

	// Tool 2: get-deals
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-deals",
		Description: "Get deals from Pipedrive with flexible filtering options including search by title, date range, owner, stage, status, and more. Use get-users tool first to find owner IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params pipedrive.GetDealsParams) (*mcp.CallToolResult, any, error) {
		result, err := client.GetDeals(ctx, params)
		if err != nil {
			return toolError("Failed to fetch deals: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccess(result)
	})

	// Tool 3: get-deal
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-deal",
		Description: "Get a specific deal by ID including custom fields",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input dealIDInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetDeal(ctx, input.DealID)
		if err != nil {
			return toolError("Failed to fetch deal %d: %s", input.DealID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 4: get-deal-notes
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-deal-notes",
		Description: "Get detailed notes and custom booking details for a specific deal",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input dealNotesInput) (*mcp.CallToolResult, any, error) {
		limit := 20
		if input.Limit != nil && *input.Limit > 0 {
			limit = *input.Limit
		}

		notes, bookingDetails, dealErr, notesErr := client.GetDealNotes(ctx, input.DealID, limit)

		result := map[string]interface{}{
			"deal_id": input.DealID,
			"notes":   notes,
		}
		if bookingDetails != nil {
			result["booking_details"] = *bookingDetails
		} else {
			result["booking_details"] = nil
		}

		if dealErr != nil {
			result["deal_error"] = pipedrive.SanitizeErrorMessage(dealErr)
		} else {
			result["deal_error"] = ""
		}
		if notesErr != nil {
			result["notes_error"] = pipedrive.SanitizeErrorMessage(notesErr)
		} else {
			result["notes_error"] = ""
		}

		result["summary"] = fmt.Sprintf("Retrieved %d notes and booking details for deal %d", len(notes), input.DealID)

		return toolSuccess(result)
	})

	// Tool 5: search-deals
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-deals",
		Description: "Search deals by term",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, any, error) {
		data, err := client.SearchDeals(ctx, input.Term)
		if err != nil {
			return toolError("Failed to search deals: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 6: get-persons
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-persons",
		Description: "Get all persons from Pipedrive including custom fields",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetPersons(ctx)
		if err != nil {
			return toolError("Failed to fetch persons: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 7: get-person
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-person",
		Description: "Get a specific person by ID including custom fields",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input personIDInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetPerson(ctx, input.PersonID)
		if err != nil {
			return toolError("Failed to fetch person %d: %s", input.PersonID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 8: search-persons
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-persons",
		Description: "Search persons by term",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, any, error) {
		data, err := client.SearchPersons(ctx, input.Term)
		if err != nil {
			return toolError("Failed to search persons: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 9: get-organizations
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-organizations",
		Description: "Get all organizations from Pipedrive including custom fields",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetOrganizations(ctx)
		if err != nil {
			return toolError("Failed to fetch organizations: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 10: get-organization
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-organization",
		Description: "Get a specific organization by ID including custom fields",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input orgIDInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetOrganization(ctx, input.OrganizationID)
		if err != nil {
			return toolError("Failed to fetch organization %d: %s", input.OrganizationID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 11: search-organizations
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-organizations",
		Description: "Search organizations by term",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, any, error) {
		data, err := client.SearchOrganizations(ctx, input.Term)
		if err != nil {
			return toolError("Failed to search organizations: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 12: get-pipelines
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-pipelines",
		Description: "Get all pipelines from Pipedrive",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetPipelines(ctx)
		if err != nil {
			return toolError("Failed to fetch pipelines: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 13: get-stages
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-stages",
		Description: "Get all stages from Pipedrive across all pipelines",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
		stages, err := client.GetStages(ctx)
		if err != nil {
			return toolError("Failed to fetch stages: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccess(stages)
	})

	// Tool 14: search-leads
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-leads",
		Description: "Search leads by term",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, any, error) {
		data, err := client.SearchLeads(ctx, input.Term)
		if err != nil {
			return toolError("Failed to search leads: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 15: search-all
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search-all",
		Description: "Search across all item types (deals, persons, organizations, etc.)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input searchAllInput) (*mcp.CallToolResult, any, error) {
		itemTypes := ""
		if input.ItemTypes != nil {
			itemTypes = *input.ItemTypes
		}
		data, err := client.SearchAll(ctx, input.Term, itemTypes)
		if err != nil {
			return toolError("Failed to search: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})
}
