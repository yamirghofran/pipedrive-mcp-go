package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
)

// --- Activity data types (v2 API) ---

// Activity represents a Pipedrive activity from the v2 API.
type Activity struct {
	ID                      int                   `json:"id"`
	Subject                 string                `json:"subject"`
	Type                    string                `json:"type"`
	OwnerID                 int                   `json:"owner_id"`
	CreatorUserID           int                   `json:"creator_user_id"`
	IsDeleted               bool                  `json:"is_deleted"`
	AddTime                 string                `json:"add_time"`
	UpdateTime              string                `json:"update_time"`
	DealID                  int                   `json:"deal_id"`
	LeadID                  string                `json:"lead_id"`
	PersonID                int                   `json:"person_id"`
	OrgID                   int                   `json:"org_id"`
	ProjectID               int                   `json:"project_id"`
	DueDate                 string                `json:"due_date"`
	DueTime                 string                `json:"due_time"`
	Duration                string                `json:"duration"`
	Busy                    bool                  `json:"busy"`
	Done                    bool                  `json:"done"`
	MarkedAsDoneTime        string                `json:"marked_as_done_time"`
	Location                *ActivityLocation     `json:"location"`
	Participants            []ActivityParticipant `json:"participants"`
	Attendees               []ActivityAttendee    `json:"attendees"`
	ConferenceMeetingClient string                `json:"conference_meeting_client"`
	ConferenceMeetingURL    string                `json:"conference_meeting_url"`
	ConferenceMeetingID     string                `json:"conference_meeting_id"`
	PublicDescription       string                `json:"public_description"`
	Priority                int                   `json:"priority"`
	Note                    string                `json:"note"`
}

// ActivityLocation represents the location of an activity.
type ActivityLocation struct {
	Value           string `json:"value,omitempty"`
	Country         string `json:"country,omitempty"`
	AdminAreaLevel1 string `json:"admin_area_level_1,omitempty"`
	AdminAreaLevel2 string `json:"admin_area_level_2,omitempty"`
	Locality        string `json:"locality,omitempty"`
	Sublocality     string `json:"sublocality,omitempty"`
	Route           string `json:"route,omitempty"`
	StreetNumber    string `json:"street_number,omitempty"`
	Subpremise      string `json:"subpremise,omitempty"`
	PostalCode      string `json:"postal_code,omitempty"`
}

// ActivityParticipant represents a participant in an activity.
type ActivityParticipant struct {
	PersonID int  `json:"person_id"`
	Primary  bool `json:"primary"`
}

// ActivityAttendee represents an attendee of an activity.
type ActivityAttendee struct {
	Email       string `json:"email,omitempty"`
	Name        string `json:"name,omitempty"`
	Status      string `json:"status,omitempty"`
	IsOrganizer bool   `json:"is_organizer,omitempty"`
	PersonID    int    `json:"person_id,omitempty"`
	UserID      int    `json:"user_id,omitempty"`
}

// --- Input parameter types ---

// GetActivitiesParams holds the parameters for listing activities.
// Uses 'any' for ID fields to flexibly accept integers, strings, or null from MCP clients.
type GetActivitiesParams struct {
	OwnerID       any `json:"ownerId,omitempty" jsonschema:"Filter by owner/user ID"`
	DealID        any `json:"dealId,omitempty" jsonschema:"Filter by deal ID"`
	LeadID        any `json:"leadId,omitempty" jsonschema:"Filter by lead ID"`
	PersonID      any `json:"personId,omitempty" jsonschema:"Filter by person ID"`
	OrgID         any `json:"orgId,omitempty" jsonschema:"Filter by organization ID"`
	Done          any `json:"done,omitempty" jsonschema:"Filter by done status (true=done, false=not done)"`
	SortBy        any `json:"sortBy,omitempty" jsonschema:"Sort field: id, update_time, add_time, due_date. Default: id"`
	SortDirection any `json:"sortDirection,omitempty" jsonschema:"Sort direction: asc or desc. Default: asc"`
	Limit         any `json:"limit,omitempty" jsonschema:"Maximum activities to return (capped at 500). Default: 100"`
	Cursor        any `json:"cursor,omitempty" jsonschema:"Pagination cursor for next page"`
}

// AddActivityParams holds the parameters for creating an activity.
// Uses 'any' for ID fields to flexibly accept integers, strings, or null from MCP clients.
type AddActivityParams struct {
	Subject           any                   `json:"subject" jsonschema:"The subject of the activity,required"`
	Type              any                   `json:"type,omitempty" jsonschema:"The type of the activity (e.g. call, meeting, task)"`
	OwnerID           any                   `json:"ownerId,omitempty" jsonschema:"ID of the user who owns the activity"`
	DealID            any                   `json:"dealId,omitempty" jsonschema:"ID of the deal to link"`
	LeadID            any                   `json:"leadId,omitempty" jsonschema:"ID of the lead to link"`
	PersonID          any                   `json:"personId,omitempty" jsonschema:"ID of the person to link"`
	OrgID             any                   `json:"orgId,omitempty" jsonschema:"ID of the organization to link"`
	ProjectID         any                   `json:"projectId,omitempty" jsonschema:"ID of the project to link"`
	DueDate           any                   `json:"dueDate,omitempty" jsonschema:"Due date in YYYY-MM-DD format"`
	DueTime           any                   `json:"dueTime,omitempty" jsonschema:"Due time in HH:MM:SS format"`
	Duration          any                   `json:"duration,omitempty" jsonschema:"Duration in HH:MM:SS format"`
	Busy              any                   `json:"busy,omitempty" jsonschema:"Whether the activity marks the assignee as busy"`
	Done              any                   `json:"done,omitempty" jsonschema:"Whether the activity is marked as done"`
	Location          *ActivityLocation     `json:"location,omitempty" jsonschema:"Location details"`
	Participants      []ActivityParticipant `json:"participants,omitempty" jsonschema:"Participants (array of {person_id, primary})"`
	PublicDescription any                   `json:"publicDescription,omitempty" jsonschema:"Public description of the activity"`
	Priority          any                   `json:"priority,omitempty" jsonschema:"Priority of the activity"`
	Note              any                   `json:"note,omitempty" jsonschema:"Note content for the activity"`
}

// UpdateActivityParams holds the parameters for updating an activity.
// Uses 'any' for ID fields to flexibly accept integers, strings, or null from MCP clients.
type UpdateActivityParams struct {
	ID                int                   `json:"-"`
	Subject           any                   `json:"subject,omitempty" jsonschema:"The subject of the activity"`
	Type              any                   `json:"type,omitempty" jsonschema:"The type of the activity"`
	OwnerID           any                   `json:"ownerId,omitempty" jsonschema:"ID of the user who owns the activity"`
	DealID            any                   `json:"dealId,omitempty" jsonschema:"ID of the deal to link"`
	LeadID            any                   `json:"leadId,omitempty" jsonschema:"ID of the lead to link"`
	PersonID          any                   `json:"personId,omitempty" jsonschema:"ID of the person to link"`
	OrgID             any                   `json:"orgId,omitempty" jsonschema:"ID of the organization to link"`
	ProjectID         any                   `json:"projectId,omitempty" jsonschema:"ID of the project to link"`
	DueDate           any                   `json:"dueDate,omitempty" jsonschema:"Due date in YYYY-MM-DD format"`
	DueTime           any                   `json:"dueTime,omitempty" jsonschema:"Due time in HH:MM:SS format"`
	Duration          any                   `json:"duration,omitempty" jsonschema:"Duration in HH:MM:SS format"`
	Busy              any                   `json:"busy,omitempty" jsonschema:"Whether the activity marks the assignee as busy"`
	Done              any                   `json:"done,omitempty" jsonschema:"Whether the activity is marked as done"`
	Location          *ActivityLocation     `json:"location,omitempty" jsonschema:"Location details"`
	Participants      []ActivityParticipant `json:"participants,omitempty" jsonschema:"Participants (array of {person_id, primary})"`
	PublicDescription any                   `json:"publicDescription,omitempty" jsonschema:"Public description of the activity"`
	Priority          any                   `json:"priority,omitempty" jsonschema:"Priority of the activity"`
	Note              any                   `json:"note,omitempty" jsonschema:"Note content for the activity"`
}

// GetActivitiesResult is the result for the get-activities tool.
type GetActivitiesResult struct {
	Activities []Activity `json:"activities"`
	NextCursor *string    `json:"next_cursor,omitempty"`
	TotalFound int        `json:"total_found"`
}

// --- Client methods ---

// GetActivities fetches activities with flexible filtering using the v2 API.
func (c *Client) GetActivities(ctx context.Context, params GetActivitiesParams) (*GetActivitiesResult, error) {
	apiParams := make(map[string]string)

	if v, err := ParseIntField(params.OwnerID); err != nil {
		return nil, fmt.Errorf("ownerId: %w", err)
	} else if v != nil {
		apiParams["owner_id"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseIntField(params.DealID); err != nil {
		return nil, fmt.Errorf("dealId: %w", err)
	} else if v != nil {
		apiParams["deal_id"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseStringField(params.LeadID); err != nil {
		return nil, fmt.Errorf("leadId: %w", err)
	} else if v != nil {
		apiParams["lead_id"] = *v
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
	if v, err := ParseBoolField(params.Done); err != nil {
		return nil, fmt.Errorf("done: %w", err)
	} else if v != nil {
		apiParams["done"] = fmt.Sprintf("%t", *v)
	}
	if v, err := ParseStringField(params.SortBy); err != nil {
		return nil, fmt.Errorf("sortBy: %w", err)
	} else if v != nil {
		apiParams["sort_by"] = *v
	}
	if v, err := ParseStringField(params.SortDirection); err != nil {
		return nil, fmt.Errorf("sortDirection: %w", err)
	} else if v != nil {
		apiParams["sort_direction"] = *v
	}
	if v, err := ParseIntField(params.Limit); err != nil {
		return nil, fmt.Errorf("limit: %w", err)
	} else if v != nil {
		apiParams["limit"] = fmt.Sprintf("%d", *v)
	}
	if v, err := ParseStringField(params.Cursor); err != nil {
		return nil, fmt.Errorf("cursor: %w", err)
	} else if v != nil && *v != "" {
		apiParams["cursor"] = *v
	}

	data, pagination, err := c.getListV2(ctx, "activities", apiParams)
	if err != nil {
		return nil, err
	}

	var activities []Activity
	if err := json.Unmarshal(data, &activities); err != nil {
		return nil, fmt.Errorf("parsing activities: %w", err)
	}

	result := &GetActivitiesResult{
		Activities: activities,
		TotalFound: len(activities),
	}

	if pagination != nil {
		result.NextCursor = pagination.NextCursor
	}

	return result, nil
}

// GetActivity fetches a specific activity by ID using the v2 API.
func (c *Client) GetActivity(ctx context.Context, activityID int) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("activities/%d", activityID)
	return c.getRawV2(ctx, endpoint, nil)
}

// AddActivity creates a new activity using the v2 API.
func (c *Client) AddActivity(ctx context.Context, params AddActivityParams) (json.RawMessage, error) {
	body := c.buildActivityRequestBody(params)
	return c.postV2(ctx, "activities", body)
}

// UpdateActivity updates an existing activity using the v2 API (PATCH).
func (c *Client) UpdateActivity(ctx context.Context, params UpdateActivityParams) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("activities/%d", params.ID)
	body := c.buildUpdateActivityRequestBody(params)
	return c.patchV2(ctx, endpoint, body)
}

// DeleteActivity deletes an activity using the v2 API.
func (c *Client) DeleteActivity(ctx context.Context, activityID int) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("activities/%d", activityID)
	return c.deleteV2(ctx, endpoint)
}

// buildActivityRequestBody converts AddActivityParams to a map for JSON submission.
func (c *Client) buildActivityRequestBody(params AddActivityParams) map[string]interface{} {
	body := make(map[string]interface{})
	if v, err := ParseStringField(params.Subject); err == nil && v != nil {
		body["subject"] = *v
	}
	if v, err := ParseStringField(params.Type); err == nil && v != nil {
		body["type"] = *v
	}
	if v, err := ParseIntField(params.OwnerID); err == nil && v != nil {
		body["owner_id"] = *v
	}
	if v, err := ParseIntField(params.DealID); err == nil && v != nil {
		body["deal_id"] = *v
	}
	if v, err := ParseStringField(params.LeadID); err == nil && v != nil {
		body["lead_id"] = *v
	}
	if v, err := ParseIntField(params.PersonID); err == nil && v != nil {
		body["person_id"] = *v
	}
	if v, err := ParseIntField(params.OrgID); err == nil && v != nil {
		body["org_id"] = *v
	}
	if v, err := ParseIntField(params.ProjectID); err == nil && v != nil {
		body["project_id"] = *v
	}
	if v, err := ParseStringField(params.DueDate); err == nil && v != nil {
		body["due_date"] = *v
	}
	if v, err := ParseStringField(params.DueTime); err == nil && v != nil {
		body["due_time"] = *v
	}
	if v, err := ParseStringField(params.Duration); err == nil && v != nil {
		body["duration"] = *v
	}
	if v, err := ParseBoolField(params.Busy); err == nil && v != nil {
		body["busy"] = *v
	}
	if v, err := ParseBoolField(params.Done); err == nil && v != nil {
		body["done"] = *v
	}
	if params.Location != nil {
		body["location"] = params.Location
	}
	if params.Participants != nil {
		body["participants"] = params.Participants
	}
	if v, err := ParseStringField(params.PublicDescription); err == nil && v != nil {
		body["public_description"] = *v
	}
	if v, err := ParseIntField(params.Priority); err == nil && v != nil {
		body["priority"] = *v
	}
	if v, err := ParseStringField(params.Note); err == nil && v != nil {
		body["note"] = *v
	}
	return body
}

// buildUpdateActivityRequestBody converts UpdateActivityParams to a map for JSON submission.
func (c *Client) buildUpdateActivityRequestBody(params UpdateActivityParams) map[string]interface{} {
	body := make(map[string]interface{})
	if v, err := ParseStringField(params.Subject); err == nil && v != nil {
		body["subject"] = *v
	}
	if v, err := ParseStringField(params.Type); err == nil && v != nil {
		body["type"] = *v
	}
	if v, err := ParseIntField(params.OwnerID); err == nil && v != nil {
		body["owner_id"] = *v
	}
	if v, err := ParseIntField(params.DealID); err == nil && v != nil {
		body["deal_id"] = *v
	}
	if v, err := ParseStringField(params.LeadID); err == nil && v != nil {
		body["lead_id"] = *v
	}
	if v, err := ParseIntField(params.PersonID); err == nil && v != nil {
		body["person_id"] = *v
	}
	if v, err := ParseIntField(params.OrgID); err == nil && v != nil {
		body["org_id"] = *v
	}
	if v, err := ParseIntField(params.ProjectID); err == nil && v != nil {
		body["project_id"] = *v
	}
	if v, err := ParseStringField(params.DueDate); err == nil && v != nil {
		body["due_date"] = *v
	}
	if v, err := ParseStringField(params.DueTime); err == nil && v != nil {
		body["due_time"] = *v
	}
	if v, err := ParseStringField(params.Duration); err == nil && v != nil {
		body["duration"] = *v
	}
	if v, err := ParseBoolField(params.Busy); err == nil && v != nil {
		body["busy"] = *v
	}
	if v, err := ParseBoolField(params.Done); err == nil && v != nil {
		body["done"] = *v
	}
	if params.Location != nil {
		body["location"] = params.Location
	}
	if params.Participants != nil {
		body["participants"] = params.Participants
	}
	if v, err := ParseStringField(params.PublicDescription); err == nil && v != nil {
		body["public_description"] = *v
	}
	if v, err := ParseIntField(params.Priority); err == nil && v != nil {
		body["priority"] = *v
	}
	if v, err := ParseStringField(params.Note); err == nil && v != nil {
		body["note"] = *v
	}
	return body
}
