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
type GetNotesParams struct {
	UserID    *int    `json:"userId,omitempty" jsonschema:"Filter by user ID (the user whose notes to fetch)"`
	DealID    *int    `json:"dealId,omitempty" jsonschema:"Filter by deal ID"`
	PersonID  *int    `json:"personId,omitempty" jsonschema:"Filter by person ID"`
	OrgID     *int    `json:"orgId,omitempty" jsonschema:"Filter by organization ID"`
	LeadID    *string `json:"leadId,omitempty" jsonschema:"Filter by lead ID (UUID)"`
	ProjectID *int    `json:"projectId,omitempty" jsonschema:"Filter by project ID"`
	Start     *int    `json:"start,omitempty" jsonschema:"Pagination start. Default: 0"`
	Limit     *int    `json:"limit,omitempty" jsonschema:"Maximum notes to return. Default: 100"`
	Sort      *string `json:"sort,omitempty" jsonschema:"Sort field and direction (e.g. 'add_time DESC')"`
}

// GetNotesResult is the result for the get-notes tool.
type GetNotesResult struct {
	Notes      []Note `json:"notes"`
	MoreItems  bool   `json:"more_items"`
	NextStart  *int   `json:"next_start,omitempty"`
	TotalFound int    `json:"total_found"`
}

// AddNoteParams holds the parameters for creating a note.
type AddNoteParams struct {
	Content   string  `json:"content" jsonschema:"Content of the note in HTML or plain text,required"`
	DealID    *int    `json:"dealId,omitempty" jsonschema:"ID of the deal to attach the note to"`
	PersonID  *int    `json:"personId,omitempty" jsonschema:"ID of the person to attach the note to"`
	OrgID     *int    `json:"orgId,omitempty" jsonschema:"ID of the organization to attach the note to"`
	LeadID    *string `json:"leadId,omitempty" jsonschema:"ID of the lead to attach the note to (UUID)"`
	ProjectID *int    `json:"projectId,omitempty" jsonschema:"ID of the project to attach the note to"`
}

// UpdateNoteParams holds the parameters for updating a note.
type UpdateNoteParams struct {
	ID        int     `json:"-"`
	Content   *string `json:"content,omitempty" jsonschema:"Updated content of the note"`
	DealID    *int    `json:"dealId,omitempty" jsonschema:"ID of the deal to attach the note to"`
	PersonID  *int    `json:"personId,omitempty" jsonschema:"ID of the person to attach the note to"`
	OrgID     *int    `json:"orgId,omitempty" jsonschema:"ID of the organization to attach the note to"`
	LeadID    *string `json:"leadId,omitempty" jsonschema:"ID of the lead to attach the note to (UUID)"`
	ProjectID *int    `json:"projectId,omitempty" jsonschema:"ID of the project to attach the note to"`
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

	if params.UserID != nil {
		apiParams["user_id"] = fmt.Sprintf("%d", *params.UserID)
	}
	if params.DealID != nil {
		apiParams["deal_id"] = fmt.Sprintf("%d", *params.DealID)
	}
	if params.PersonID != nil {
		apiParams["person_id"] = fmt.Sprintf("%d", *params.PersonID)
	}
	if params.OrgID != nil {
		apiParams["org_id"] = fmt.Sprintf("%d", *params.OrgID)
	}
	if params.LeadID != nil {
		apiParams["lead_id"] = *params.LeadID
	}
	if params.ProjectID != nil {
		apiParams["project_id"] = fmt.Sprintf("%d", *params.ProjectID)
	}
	if params.Start != nil {
		apiParams["start"] = fmt.Sprintf("%d", *params.Start)
	}
	if params.Limit != nil {
		apiParams["limit"] = fmt.Sprintf("%d", *params.Limit)
	}
	if params.Sort != nil {
		apiParams["sort"] = *params.Sort
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
	if params.DealID != nil {
		body["deal_id"] = *params.DealID
	}
	if params.PersonID != nil {
		body["person_id"] = *params.PersonID
	}
	if params.OrgID != nil {
		body["org_id"] = *params.OrgID
	}
	if params.LeadID != nil {
		body["lead_id"] = *params.LeadID
	}
	if params.ProjectID != nil {
		body["project_id"] = *params.ProjectID
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
	if params.DealID != nil {
		body["deal_id"] = *params.DealID
	}
	if params.PersonID != nil {
		body["person_id"] = *params.PersonID
	}
	if params.OrgID != nil {
		body["org_id"] = *params.OrgID
	}
	if params.LeadID != nil {
		body["lead_id"] = *params.LeadID
	}
	if params.ProjectID != nil {
		body["project_id"] = *params.ProjectID
	}

	return c.putV1(ctx, endpoint, body)
}

// DeleteNote deletes a note using the v1 API.
func (c *Client) DeleteNote(ctx context.Context, noteID int) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("notes/%d", noteID)
	return c.deleteV1(ctx, endpoint)
}
