package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
)

// Note represents a Pipedrive note from the v1 API.
type Note struct {
	ID                       int             `json:"id"`
	ActiveFlag               bool            `json:"active_flag"`
	AddTime                  string          `json:"add_time"`
	UpdateTime               string          `json:"update_time"`
	Content                  string          `json:"content"`
	DealID                   int             `json:"deal_id"`
	LeadID                   string          `json:"lead_id"`
	PersonID                 int             `json:"person_id"`
	OrgID                    int             `json:"org_id"`
	ProjectID                int             `json:"project_id"`
	LastUpdateUserID         int             `json:"last_update_user_id"`
	UserID                   int             `json:"user_id"`
	PinnedToDealFlag         bool            `json:"pinned_to_deal_flag"`
	PinnedToOrganizationFlag bool            `json:"pinned_to_organization_flag"`
	PinnedToPersonFlag       bool            `json:"pinned_to_person_flag"`
	PinnedToProjectFlag      bool            `json:"pinned_to_project_flag"`
	Raw                      json.RawMessage `json:"-"`
}

// GetNotesParams holds the parameters for listing notes.
// Uses 'any' for ID fields to flexibly accept integers, strings, or null from MCP clients.
type GetNotesParams struct {
	UserID    any `json:"userId,omitempty" jsonschema:"Filter by user ID (the user whose notes to fetch)"`
	DealID    any `json:"dealId,omitempty" jsonschema:"Filter by deal ID"`
	PersonID  any `json:"personId,omitempty" jsonschema:"Filter by person ID"`
	OrgID     any `json:"orgId,omitempty" jsonschema:"Filter by organization ID"`
	LeadID    any `json:"leadId,omitempty" jsonschema:"Filter by lead ID (UUID)"`
	ProjectID any `json:"projectId,omitempty" jsonschema:"Filter by project ID"`
	Start     any `json:"start,omitempty" jsonschema:"Pagination start. Default: 0"`
	Limit     any `json:"limit,omitempty" jsonschema:"Maximum notes to return. Default: 100"`
	Sort      any `json:"sort,omitempty" jsonschema:"Sort field and direction (e.g. 'add_time DESC')"`
}

// GetNotesResult is the result for the get-notes tool.
type GetNotesResult struct {
	Notes      []Note `json:"notes"`
	MoreItems  bool   `json:"more_items"`
	NextStart  *int   `json:"next_start,omitempty"`
	TotalFound int    `json:"total_found"`
}

// AddNoteParams holds the parameters for creating a note.
// Uses 'any' for ID fields to flexibly accept integers, strings, or null from MCP clients.
type AddNoteParams struct {
	Content   string `json:"content" jsonschema:"Content of the note in HTML or plain text,required"`
	DealID    any    `json:"dealId,omitempty" jsonschema:"ID of the deal to attach the note to"`
	PersonID  any    `json:"personId,omitempty" jsonschema:"ID of the person to attach the note to"`
	OrgID     any    `json:"orgId,omitempty" jsonschema:"ID of the organization to attach the note to"`
	LeadID    any    `json:"leadId,omitempty" jsonschema:"ID of the lead to attach the note to (UUID)"`
	ProjectID any    `json:"projectId,omitempty" jsonschema:"ID of the project to attach the note to"`
}

// UpdateNoteParams holds the parameters for updating a note.
// Uses 'any' for ID fields to flexibly accept integers, strings, or null from MCP clients.
type UpdateNoteParams struct {
	ID        int     `json:"-"`
	Content   *string `json:"content,omitempty" jsonschema:"Updated content of the note"`
	DealID    any     `json:"dealId,omitempty" jsonschema:"ID of the deal to attach the note to"`
	PersonID  any     `json:"personId,omitempty" jsonschema:"ID of the person to attach the note to"`
	OrgID     any     `json:"orgId,omitempty" jsonschema:"ID of the organization to attach the note to"`
	LeadID    any     `json:"leadId,omitempty" jsonschema:"ID of the lead to attach the note to (UUID)"`
	ProjectID any     `json:"projectId,omitempty" jsonschema:"ID of the project to attach the note to"`
}

// GetDealNotes fetches notes for a specific deal, plus the deal's booking details.
func (c *Client) GetDealNotes(ctx context.Context, dealID int, limit int) (notes []Note, bookingDetails *string, dealErr error, notesErr error) {
	// Fetch deal and notes concurrently using goroutines
	type dealResult struct {
		data json.RawMessage
		err  error
	}
	type notesResult struct {
		data json.RawMessage
		err  error
	}

	dealCh := make(chan dealResult, 1)
	notesCh := make(chan notesResult, 1)

	// Fetch deal details
	go func() {
		data, err := c.getRaw(ctx, fmt.Sprintf("deals/%d", dealID), nil)
		dealCh <- dealResult{data: data, err: err}
	}()

	// Fetch notes
	go func() {
		data, _, err := c.getList(ctx, "notes", map[string]string{
			"deal_id": fmt.Sprintf("%d", dealID),
			"limit":   fmt.Sprintf("%d", limit),
		})
		notesCh <- notesResult{data: data, err: err}
	}()

	// Process deal response
	dr := <-dealCh
	if dr.err != nil {
		dealErr = dr.err
	} else {
		var dealMap map[string]interface{}
		if err := json.Unmarshal(dr.data, &dealMap); err == nil {
			if v, ok := dealMap[c.bookingFieldKey]; ok {
				if s, ok := v.(string); ok && s != "" {
					bookingDetails = &s
				}
			}
		}
	}

	// Process notes response
	nr := <-notesCh
	if nr.err != nil {
		notesErr = nr.err
	} else {
		if err := json.Unmarshal(nr.data, &notes); err != nil {
			notesErr = err
		}
	}

	return notes, bookingDetails, dealErr, notesErr
}

// GetNotes fetches notes with flexible filtering using the v1 API.
func (c *Client) GetNotes(ctx context.Context, params GetNotesParams) (*GetNotesResult, error) {
	apiParams := make(map[string]string)

	if v, err := ParseIntField(params.UserID); err != nil {
		return nil, fmt.Errorf("userId: %w", err)
	} else if v != nil {
		apiParams["user_id"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseIntField(params.DealID); err != nil {
		return nil, fmt.Errorf("dealId: %w", err)
	} else if v != nil {
		apiParams["deal_id"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseIntField(params.PersonID); err != nil {
		return nil, fmt.Errorf("personId: %w", err)
	} else if v != nil {
		apiParams["person_id"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseIntField(params.OrgID); err != nil {
		return nil, fmt.Errorf("orgId: %w", err)
	} else if v != nil {
		apiParams["org_id"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseStringField(params.LeadID); err != nil {
		return nil, fmt.Errorf("leadId: %w", err)
	} else if v != nil {
		apiParams["lead_id"] = *v
	}
	if v, err := ParseIntField(params.ProjectID); err != nil {
		return nil, fmt.Errorf("projectId: %w", err)
	} else if v != nil {
		apiParams["project_id"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseIntField(params.Start); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	} else if v != nil {
		apiParams["start"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseIntField(params.Limit); err != nil {
		return nil, fmt.Errorf("limit: %w", err)
	} else if v != nil {
		apiParams["limit"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseStringField(params.Sort); err != nil {
		return nil, fmt.Errorf("sort: %w", err)
	} else if v != nil {
		apiParams["sort"] = *v
	}

	data, pagination, err := c.getList(ctx, "notes", apiParams)
	if err != nil {
		return nil, err
	}

	var notes []Note
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, fmt.Errorf("parsing notes: %w", err)
	}

	result := &GetNotesResult{
		Notes:      notes,
		TotalFound: len(notes),
	}

	if pagination != nil && pagination.Pagination.MoreItemsInCollection {
		result.MoreItems = true
		nextStart := pagination.Pagination.Start + pagination.Pagination.Limit
		result.NextStart = &nextStart
	}

	return result, nil
}

// GetNote fetches a specific note by ID.
func (c *Client) GetNote(ctx context.Context, noteID int) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("notes/%d", noteID)
	return c.getRaw(ctx, endpoint, nil)
}

// AddNote creates a new note using the v1 API.
func (c *Client) AddNote(ctx context.Context, params AddNoteParams) (json.RawMessage, error) {
	body := map[string]interface{}{
		"content": params.Content,
	}
	if v, err := ParseIntField(params.DealID); err != nil {
		return nil, fmt.Errorf("dealId: %w", err)
	} else if v != nil {
		body["deal_id"] = *v
	}
	if v, err := ParseIntField(params.PersonID); err != nil {
		return nil, fmt.Errorf("personId: %w", err)
	} else if v != nil {
		body["person_id"] = *v
	}
	if v, err := ParseIntField(params.OrgID); err != nil {
		return nil, fmt.Errorf("orgId: %w", err)
	} else if v != nil {
		body["org_id"] = *v
	}
	if v, err := ParseStringField(params.LeadID); err != nil {
		return nil, fmt.Errorf("leadId: %w", err)
	} else if v != nil {
		body["lead_id"] = *v
	}
	if v, err := ParseIntField(params.ProjectID); err != nil {
		return nil, fmt.Errorf("projectId: %w", err)
	} else if v != nil {
		body["project_id"] = *v
	}

	return c.postV1(ctx, "notes", body)
}

// UpdateNote updates an existing note using the v1 API (PUT).
func (c *Client) UpdateNote(ctx context.Context, params UpdateNoteParams) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("notes/%d", params.ID)
	body := make(map[string]interface{})
	if params.Content != nil {
		body["content"] = *params.Content
	}
	if v, err := ParseIntField(params.DealID); err != nil {
		return nil, fmt.Errorf("dealId: %w", err)
	} else if v != nil {
		body["deal_id"] = *v
	}
	if v, err := ParseIntField(params.PersonID); err != nil {
		return nil, fmt.Errorf("personId: %w", err)
	} else if v != nil {
		body["person_id"] = *v
	}
	if v, err := ParseIntField(params.OrgID); err != nil {
		return nil, fmt.Errorf("orgId: %w", err)
	} else if v != nil {
		body["org_id"] = *v
	}
	if v, err := ParseStringField(params.LeadID); err != nil {
		return nil, fmt.Errorf("leadId: %w", err)
	} else if v != nil {
		body["lead_id"] = *v
	}
	if v, err := ParseIntField(params.ProjectID); err != nil {
		return nil, fmt.Errorf("projectId: %w", err)
	} else if v != nil {
		body["project_id"] = *v
	}

	return c.putV1(ctx, endpoint, body)
}

// DeleteNote deletes a note using the v1 API.
func (c *Client) DeleteNote(ctx context.Context, noteID int) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("notes/%d", noteID)
	return c.deleteV1(ctx, endpoint)
}
