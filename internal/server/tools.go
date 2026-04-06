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

type activityIDInput struct {
	ActivityID int `json:"activityId" jsonschema:"Pipedrive activity ID,required"`
}

type noteIDInput struct {
	NoteID int `json:"noteId" jsonschema:"Pipedrive note ID,required"`
}

// registerTools registers all MCP tools.
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
		Description: "List and filter deals from Pipedrive. This is the primary tool for fetching deals — use it when the user asks to see deals, list deals, or get an overview of their pipeline. Supports filtering by title search, status (open/won/lost/deleted), owner, stage, pipeline, value range, and date range. By default returns all non-deleted deals updated in the last 365 days. Use get-users tool first to find owner IDs.",
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
		Description: "Search Pipedrive leads (not deals) by term. Requires a search term of at least 2 characters. Do NOT use this tool to list or fetch deals — use get-deals instead.",
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

	// --- Activity Tools (v2 API) ---

	// Tool 16: get-activities
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-activities",
		Description: "Get activities from Pipedrive with flexible filtering by owner, deal, person, organization, or done status. Supports pagination with cursor.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params pipedrive.GetActivitiesParams) (*mcp.CallToolResult, any, error) {
		result, err := client.GetActivities(ctx, params)
		if err != nil {
			return toolError("Failed to fetch activities: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccess(result)
	})

	// Tool 17: get-activity
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-activity",
		Description: "Get a specific activity by ID with full details including location, participants, and conference info",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input activityIDInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetActivity(ctx, input.ActivityID)
		if err != nil {
			return toolError("Failed to fetch activity %d: %s", input.ActivityID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 18: add-activity
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add-activity",
		Description: "Create a new activity in Pipedrive. Requires a subject. Optionally link to a deal, person, organization, or lead. Set due date, duration, type, and more.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params pipedrive.AddActivityParams) (*mcp.CallToolResult, any, error) {
		if params.Subject == nil || *params.Subject == "" {
			return toolError("Subject is required to create an activity")
		}
		data, err := client.AddActivity(ctx, params)
		if err != nil {
			return toolError("Failed to create activity: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 19: update-activity
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-activity",
		Description: "Update an existing activity. Can mark as done, change subject, type, due date, assignee, or any other field.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params pipedrive.UpdateActivityParams) (*mcp.CallToolResult, any, error) {
		if params.ID == 0 {
			return toolError("Activity ID is required to update an activity")
		}
		data, err := client.UpdateActivity(ctx, params)
		if err != nil {
			return toolError("Failed to update activity %d: %s", params.ID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 20: delete-activity
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-activity",
		Description: "Delete an activity by ID. The activity is marked as deleted and permanently removed after 30 days.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input activityIDInput) (*mcp.CallToolResult, any, error) {
		data, err := client.DeleteActivity(ctx, input.ActivityID)
		if err != nil {
			return toolError("Failed to delete activity %d: %s", input.ActivityID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccess(map[string]interface{}{
			"summary":     fmt.Sprintf("Activity %d deleted", input.ActivityID),
			"activity_id": input.ActivityID,
			"deleted":     true,
			"details":     data,
		})
	})

	// --- Note Tools (v1 API) ---

	// Tool 21: get-notes
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-notes",
		Description: "Get notes from Pipedrive with flexible filtering by deal, person, organization, user, or lead. Supports pagination.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params pipedrive.GetNotesParams) (*mcp.CallToolResult, any, error) {
		result, err := client.GetNotes(ctx, params)
		if err != nil {
			return toolError("Failed to fetch notes: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccess(result)
	})

	// Tool 22: get-note
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get-note",
		Description: "Get a specific note by ID with full details including content, associated entities, and creator info",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input noteIDInput) (*mcp.CallToolResult, any, error) {
		data, err := client.GetNote(ctx, input.NoteID)
		if err != nil {
			return toolError("Failed to fetch note %d: %s", input.NoteID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 23: add-note
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add-note",
		Description: "Create a new note in Pipedrive. Requires content and at least one associated entity (deal, person, organization, lead, or project).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params pipedrive.AddNoteParams) (*mcp.CallToolResult, any, error) {
		if params.Content == "" {
			return toolError("Content is required to create a note")
		}
		if params.DealID == nil && params.PersonID == nil && params.OrgID == nil && params.LeadID == nil && params.ProjectID == nil {
			return toolError("At least one associated entity (dealId, personId, orgId, leadId, or projectId) is required")
		}
		data, err := client.AddNote(ctx, params)
		if err != nil {
			return toolError("Failed to create note: %s", pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 24: update-note
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update-note",
		Description: "Update an existing note's content or reassign it to a different deal, person, or organization.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params pipedrive.UpdateNoteParams) (*mcp.CallToolResult, any, error) {
		if params.ID == 0 {
			return toolError("Note ID is required to update a note")
		}
		data, err := client.UpdateNote(ctx, params)
		if err != nil {
			return toolError("Failed to update note %d: %s", params.ID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccessRaw(data)
	})

	// Tool 25: delete-note
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete-note",
		Description: "Delete a note by ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input noteIDInput) (*mcp.CallToolResult, any, error) {
		data, err := client.DeleteNote(ctx, input.NoteID)
		if err != nil {
			return toolError("Failed to delete note %d: %s", input.NoteID, pipedrive.SanitizeErrorMessage(err))
		}
		return toolSuccess(map[string]interface{}{
			"summary": fmt.Sprintf("Note %d deleted", input.NoteID),
			"note_id": input.NoteID,
			"deleted": true,
			"details": data,
		})
	})
}
