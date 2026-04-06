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
type GetActivitiesParams struct {
	OwnerID       *int    `json:"ownerId,omitempty" jsonschema:"Filter by owner/user ID"`
	DealID        *int    `json:"dealId,omitempty" jsonschema:"Filter by deal ID"`
	LeadID        *string `json:"leadId,omitempty" jsonschema:"Filter by lead ID"`
	PersonID      *int    `json:"personId,omitempty" jsonschema:"Filter by person ID"`
	OrgID         *int    `json:"orgId,omitempty" jsonschema:"Filter by organization ID"`
	Done          *bool   `json:"done,omitempty" jsonschema:"Filter by done status (true=done, false=not done)"`
	SortBy        *string `json:"sortBy,omitempty" jsonschema:"Sort field: id, update_time, add_time, due_date. Default: id"`
	SortDirection *string `json:"sortDirection,omitempty" jsonschema:"Sort direction: asc or desc. Default: asc"`
	Limit         *int    `json:"limit,omitempty" jsonschema:"Maximum activities to return (capped at 500). Default: 100"`
	Cursor        *string `json:"cursor,omitempty" jsonschema:"Pagination cursor for next page"`
}

// AddActivityParams holds the parameters for creating an activity.
type AddActivityParams struct {
	Subject           *string               `json:"subject" jsonschema:"The subject of the activity,required"`
	Type              *string               `json:"type,omitempty" jsonschema:"The type of the activity (e.g. call, meeting, task)"`
	OwnerID           *int                  `json:"ownerId,omitempty" jsonschema:"ID of the user who owns the activity"`
	DealID            *int                  `json:"dealId,omitempty" jsonschema:"ID of the deal to link"`
	LeadID            *string               `json:"leadId,omitempty" jsonschema:"ID of the lead to link"`
	PersonID          *int                  `json:"personId,omitempty" jsonschema:"ID of the person to link"`
	OrgID             *int                  `json:"orgId,omitempty" jsonschema:"ID of the organization to link"`
	ProjectID         *int                  `json:"projectId,omitempty" jsonschema:"ID of the project to link"`
	DueDate           *string               `json:"dueDate,omitempty" jsonschema:"Due date in YYYY-MM-DD format"`
	DueTime           *string               `json:"dueTime,omitempty" jsonschema:"Due time in HH:MM:SS format"`
	Duration          *string               `json:"duration,omitempty" jsonschema:"Duration in HH:MM:SS format"`
	Busy              *bool                 `json:"busy,omitempty" jsonschema:"Whether the activity marks the assignee as busy"`
	Done              *bool                 `json:"done,omitempty" jsonschema:"Whether the activity is marked as done"`
	Location          *ActivityLocation     `json:"location,omitempty" jsonschema:"Location details"`
	Participants      []ActivityParticipant `json:"participants,omitempty" jsonschema:"Participants (array of {person_id, primary})"`
	PublicDescription *string               `json:"publicDescription,omitempty" jsonschema:"Public description of the activity"`
	Priority          *int                  `json:"priority,omitempty" jsonschema:"Priority of the activity"`
	Note              *string               `json:"note,omitempty" jsonschema:"Note content for the activity"`
}

// UpdateActivityParams holds the parameters for updating an activity.
type UpdateActivityParams struct {
	ID                int                   `json:"-"`
	Subject           *string               `json:"subject,omitempty" jsonschema:"The subject of the activity"`
	Type              *string               `json:"type,omitempty" jsonschema:"The type of the activity"`
	OwnerID           *int                  `json:"ownerId,omitempty" jsonschema:"ID of the user who owns the activity"`
	DealID            *int                  `json:"dealId,omitempty" jsonschema:"ID of the deal to link"`
	LeadID            *string               `json:"leadId,omitempty" jsonschema:"ID of the lead to link"`
	PersonID          *int                  `json:"personId,omitempty" jsonschema:"ID of the person to link"`
	OrgID             *int                  `json:"orgId,omitempty" jsonschema:"ID of the organization to link"`
	ProjectID         *int                  `json:"projectId,omitempty" jsonschema:"ID of the project to link"`
	DueDate           *string               `json:"dueDate,omitempty" jsonschema:"Due date in YYYY-MM-DD format"`
	DueTime           *string               `json:"dueTime,omitempty" jsonschema:"Due time in HH:MM:SS format"`
	Duration          *string               `json:"duration,omitempty" jsonschema:"Duration in HH:MM:SS format"`
	Busy              *bool                 `json:"busy,omitempty" jsonschema:"Whether the activity marks the assignee as busy"`
	Done              *bool                 `json:"done,omitempty" jsonschema:"Whether the activity is marked as done"`
	Location          *ActivityLocation     `json:"location,omitempty" jsonschema:"Location details"`
	Participants      []ActivityParticipant `json:"participants,omitempty" jsonschema:"Participants (array of {person_id, primary})"`
	PublicDescription *string               `json:"publicDescription,omitempty" jsonschema:"Public description of the activity"`
	Priority          *int                  `json:"priority,omitempty" jsonschema:"Priority of the activity"`
	Note              *string               `json:"note,omitempty" jsonschema:"Note content for the activity"`
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

	if params.OwnerID != nil {
		apiParams["owner_id"] = fmt.Sprintf("%d", *params.OwnerID)
	}
	if params.DealID != nil {
		apiParams["deal_id"] = fmt.Sprintf("%d", *params.DealID)
	}
	if params.LeadID != nil {
		apiParams["lead_id"] = *params.LeadID
	}
	if params.PersonID != nil {
		apiParams["person_id"] = fmt.Sprintf("%d", *params.PersonID)
	}
	if params.OrgID != nil {
		apiParams["org_id"] = fmt.Sprintf("%d", *params.OrgID)
	}
	if params.Done != nil {
		apiParams["done"] = fmt.Sprintf("%t", *params.Done)
	}
	if params.SortBy != nil {
		apiParams["sort_by"] = *params.SortBy
	}
	if params.SortDirection != nil {
		apiParams["sort_direction"] = *params.SortDirection
	}
	if params.Limit != nil {
		apiParams["limit"] = fmt.Sprintf("%d", *params.Limit)
	}
	if params.Cursor != nil && *params.Cursor != "" {
		apiParams["cursor"] = *params.Cursor
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
	if params.Subject != nil {
		body["subject"] = *params.Subject
	}
	if params.Type != nil {
		body["type"] = *params.Type
	}
	if params.OwnerID != nil {
		body["owner_id"] = *params.OwnerID
	}
	if params.DealID != nil {
		body["deal_id"] = *params.DealID
	}
	if params.LeadID != nil {
		body["lead_id"] = *params.LeadID
	}
	if params.PersonID != nil {
		body["person_id"] = *params.PersonID
	}
	if params.OrgID != nil {
		body["org_id"] = *params.OrgID
	}
	if params.ProjectID != nil {
		body["project_id"] = *params.ProjectID
	}
	if params.DueDate != nil {
		body["due_date"] = *params.DueDate
	}
	if params.DueTime != nil {
		body["due_time"] = *params.DueTime
	}
	if params.Duration != nil {
		body["duration"] = *params.Duration
	}
	if params.Busy != nil {
		body["busy"] = *params.Busy
	}
	if params.Done != nil {
		body["done"] = *params.Done
	}
	if params.Location != nil {
		body["location"] = params.Location
	}
	if params.Participants != nil {
		body["participants"] = params.Participants
	}
	if params.PublicDescription != nil {
		body["public_description"] = *params.PublicDescription
	}
	if params.Priority != nil {
		body["priority"] = *params.Priority
	}
	if params.Note != nil {
		body["note"] = *params.Note
	}
	return body
}

// buildUpdateActivityRequestBody converts UpdateActivityParams to a map for JSON submission.
func (c *Client) buildUpdateActivityRequestBody(params UpdateActivityParams) map[string]interface{} {
	body := make(map[string]interface{})
	if params.Subject != nil {
		body["subject"] = *params.Subject
	}
	if params.Type != nil {
		body["type"] = *params.Type
	}
	if params.OwnerID != nil {
		body["owner_id"] = *params.OwnerID
	}
	if params.DealID != nil {
		body["deal_id"] = *params.DealID
	}
	if params.LeadID != nil {
		body["lead_id"] = *params.LeadID
	}
	if params.PersonID != nil {
		body["person_id"] = *params.PersonID
	}
	if params.OrgID != nil {
		body["org_id"] = *params.OrgID
	}
	if params.ProjectID != nil {
		body["project_id"] = *params.ProjectID
	}
	if params.DueDate != nil {
		body["due_date"] = *params.DueDate
	}
	if params.DueTime != nil {
		body["due_time"] = *params.DueTime
	}
	if params.Duration != nil {
		body["duration"] = *params.Duration
	}
	if params.Busy != nil {
		body["busy"] = *params.Busy
	}
	if params.Done != nil {
		body["done"] = *params.Done
	}
	if params.Location != nil {
		body["location"] = params.Location
	}
	if params.Participants != nil {
		body["participants"] = params.Participants
	}
	if params.PublicDescription != nil {
		body["public_description"] = *params.PublicDescription
	}
	if params.Priority != nil {
		body["priority"] = *params.Priority
	}
	if params.Note != nil {
		body["note"] = *params.Note
	}
	return body
}
