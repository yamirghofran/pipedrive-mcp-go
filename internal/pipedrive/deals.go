package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// --- Data types for deal-related responses ---

// RawDeal represents a raw deal from the Pipedrive list deals endpoint.
type RawDeal struct {
	ID               int           `json:"id"`
	Title            string        `json:"title"`
	Value            float64       `json:"value"`
	Currency         string        `json:"currency"`
	Status           string        `json:"status"`
	StageID          int           `json:"stage_id"`
	PipelineID       int           `json:"pipeline_id"`
	UserID           json.Number   `json:"user_id"`
	AddTime          string        `json:"add_time"`
	UpdateTime       string        `json:"update_time"`
	LastActivityDate string        `json:"last_activity_date"`
	CloseTime        string        `json:"close_time"`
	WonTime          string        `json:"won_time"`
	LostTime         string        `json:"lost_time"`
	NotesCount       int           `json:"notes_count"`
	OrgID            *NestedIDName `json:"org_id"`
	PersonID         *NestedIDName `json:"person_id"`
	StageName        string        `json:"stage_name"`    // from search results
	PipelineName     string        `json:"pipeline_name"` // from search results
	OwnerName        string        `json:"owner_name"`    // from search results
	StageOrderNr     int           `json:"stage_order_nr"`
	// Custom fields stored as raw JSON
	CustomAttributes json.RawMessage `json:"-"`
}

// NestedIDName represents a nested object with id and name fields.
type NestedIDName struct {
	Value interface{} `json:"value"`
	Name  string      `json:"name"`
}

// GetName safely returns the name from a NestedIDName.
func (n *NestedIDName) GetName() string {
	if n == nil {
		return ""
	}
	return n.Name
}

// RawDealSearchResultItem represents an item from the deal search endpoint.
type RawDealSearchResultItem struct {
	ID           int         `json:"id"`
	Title        string      `json:"title"`
	Value        float64     `json:"value"`
	Currency     string      `json:"currency"`
	Status       string      `json:"status"`
	StageID      int         `json:"stage_id"`
	PipelineID   int         `json:"pipeline_id"`
	UserID       json.Number `json:"user_id"`
	AddTime      string      `json:"add_time"`
	WonTime      string      `json:"won_time"`
	LostTime     string      `json:"lost_time"`
	CloseTime    string      `json:"close_time"`
	Organization *struct {
		Name string `json:"name"`
	} `json:"organization"`
	Person *struct {
		Name string `json:"name"`
	} `json:"person"`
	Owner *struct {
		Name string `json:"name"`
	} `json:"owner"`
	Stage *struct {
		Name string `json:"name"`
	} `json:"stage"`
	Pipeline *struct {
		Name string `json:"name"`
	} `json:"pipeline"`
}

// DealSummary is the summarized deal returned in MCP tool responses.
type DealSummary struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	Value            float64 `json:"value"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
	StageName        string  `json:"stage_name"`
	PipelineName     string  `json:"pipeline_name"`
	OwnerName        string  `json:"owner_name"`
	OrganizationName *string `json:"_name"`
	PersonName       *string `json:"person_name"`
	AddTime          string  `json:"add_time"`
	LastActivityDate string  `json:"last_activity_date"`
	CloseTime        string  `json:"close_time"`
	WonTime          string  `json:"won_time"`
	LostTime         string  `json:"lost_time"`
	NotesCount       int     `json:"notes_count"`
	Notes            []Note  `json:"notes"`
	BookingDetails   *string `json:"booking_details"`
}

// GetDealsParams holds the parameters for the get-deals tool.
type GetDealsParams struct {
	SearchTitle *string  `json:"searchTitle,omitempty" jsonschema:"Search deals by title (partial matches)"`
	DaysBack    *int     `json:"daysBack,omitempty" jsonschema:"Days back to fetch based on last activity date. Default: 365"`
	OwnerID     *int     `json:"ownerId,omitempty" jsonschema:"Filter by owner/user ID"`
	StageID     *int     `json:"stageId,omitempty" jsonschema:"Filter by stage ID"`
	Status      *string  `json:"status,omitempty" jsonschema:"Deal status: open, won, lost, or deleted. Default: open"`
	PipelineID  *int     `json:"pipelineId,omitempty" jsonschema:"Filter by pipeline ID"`
	MinValue    *float64 `json:"minValue,omitempty" jsonschema:"Minimum deal value"`
	MaxValue    *float64 `json:"maxValue,omitempty" jsonschema:"Maximum deal value"`
	Limit       *int     `json:"limit,omitempty" jsonschema:"Maximum deals to return (capped at 500). Default: 500"`
}

// GetDealsResult is the response for the get-deals tool.
type GetDealsResult struct {
	Summary        string                 `json:"summary"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
	TotalFound     int                    `json:"total_found"`
	Deals          []DealSummary          `json:"deals"`
}

// GetDeals fetches deals with flexible filtering.
func (c *Client) GetDeals(ctx context.Context, params GetDealsParams) (*GetDealsResult, error) {
	// Apply defaults
	daysBack := 365
	if params.DaysBack != nil {
		daysBack = *params.DaysBack
	}
	status := "open"
	if params.Status != nil && *params.Status != "" {
		status = *params.Status
	}
	limit := 500
	if params.Limit != nil {
		if *params.Limit < 500 {
			limit = *params.Limit
		}
	}

	var dealSummaries []DealSummary
	var filtersApplied = make(map[string]interface{})

	if params.SearchTitle != nil && *params.SearchTitle != "" {
		// Search path
		filtersApplied["searchTitle"] = *params.SearchTitle
		searchResults, err := c.searchDealsRaw(ctx, *params.SearchTitle)
		if err != nil {
			return nil, err
		}

		dealSummaries = c.filterAndSummarizeSearchResults(searchResults, params, limit)
	} else {
		// List path
		apiParams := map[string]string{
			"sort":   "last_activity_date DESC",
			"status": status,
			"limit":  fmt.Sprintf("%d", limit),
		}
		if params.OwnerID != nil {
			apiParams["user_id"] = fmt.Sprintf("%d", *params.OwnerID)
			filtersApplied["ownerId"] = *params.OwnerID
		}
		if params.StageID != nil {
			apiParams["stage_id"] = fmt.Sprintf("%d", *params.StageID)
			filtersApplied["stageId"] = *params.StageID
		}
		if params.PipelineID != nil {
			apiParams["pipeline_id"] = fmt.Sprintf("%d", *params.PipelineID)
			filtersApplied["pipelineId"] = *params.PipelineID
		}

		if status != "open" {
			filtersApplied["status"] = status
		}

		data, _, err := c.getList(ctx, "deals", apiParams)
		if err != nil {
			return nil, err
		}

		dealSummaries = c.filterAndSummarizeListedDeals(data, daysBack, params, limit)
	}

	// Track additional filters
	if params.MinValue != nil {
		filtersApplied["minValue"] = *params.MinValue
	}
	if params.MaxValue != nil {
		filtersApplied["maxValue"] = *params.MaxValue
	}
	if params.DaysBack != nil {
		filtersApplied["daysBack"] = *params.DaysBack
	}

	totalFound := len(dealSummaries)

	// Cap at 30 for response
	if len(dealSummaries) > 30 {
		dealSummaries = dealSummaries[:30]
	}

	summary := fmt.Sprintf("Found %d deals matching your criteria", totalFound)
	if totalFound > 30 {
		summary = fmt.Sprintf("Found %d deals matching your criteria (showing first 30)", totalFound)
	}

	return &GetDealsResult{
		Summary:        summary,
		FiltersApplied: filtersApplied,
		TotalFound:     totalFound,
		Deals:          dealSummaries,
	}, nil
}

// searchDealsRaw searches deals by term using the search endpoint.
func (c *Client) searchDealsRaw(ctx context.Context, term string) ([]RawDealSearchResultItem, error) {
	data, err := c.getRaw(ctx, "deals/search", map[string]string{"term": term})
	if err != nil {
		return nil, err
	}

	var items []RawDealSearchResultItem
	if err := json.Unmarshal(data, &items); err != nil {
		// Try as a wrapper object
		var wrapper struct {
			Items []RawDealSearchResultItem `json:"items"`
		}
		if err2 := json.Unmarshal(data, &wrapper); err2 != nil {
			return nil, fmt.Errorf("parsing deal search results: %w", err)
		}
		return wrapper.Items, nil
	}
	return items, nil
}

// filterAndSummarizeSearchResults applies client-side filters to search results.
func (c *Client) filterAndSummarizeSearchResults(items []RawDealSearchResultItem, params GetDealsParams, limit int) []DealSummary {
	var results []DealSummary

	for _, item := range items {
		// Apply owner filter
		if params.OwnerID != nil {
			uid, _ := item.UserID.Int64()
			if int(uid) != *params.OwnerID {
				continue
			}
		}

		// Apply status filter
		if params.Status != nil && *params.Status != "" && item.Status != *params.Status {
			continue
		}

		// Apply stage filter
		if params.StageID != nil && item.StageID != *params.StageID {
			continue
		}

		// Apply pipeline filter
		if params.PipelineID != nil && item.PipelineID != *params.PipelineID {
			continue
		}

		// Apply min/max value filters
		if params.MinValue != nil && item.Value < *params.MinValue {
			continue
		}
		if params.MaxValue != nil && item.Value > *params.MaxValue {
			continue
		}

		stageName := "Unknown"
		if item.Stage != nil {
			stageName = item.Stage.Name
		}
		pipelineName := "Unknown"
		if item.Pipeline != nil {
			pipelineName = item.Pipeline.Name
		}
		ownerName := "Unknown"
		if item.Owner != nil {
			ownerName = item.Owner.Name
		}

		var orgName *string
		if item.Organization != nil && item.Organization.Name != "" {
			orgName = &item.Organization.Name
		}
		var personName *string
		if item.Person != nil && item.Person.Name != "" {
			personName = &item.Person.Name
		}

		summary := DealSummary{
			ID:               item.ID,
			Title:            item.Title,
			Value:            item.Value,
			Currency:         item.Currency,
			Status:           item.Status,
			StageName:        stageName,
			PipelineName:     pipelineName,
			OwnerName:        ownerName,
			OrganizationName: orgName,
			PersonName:       personName,
			AddTime:          item.AddTime,
			CloseTime:        item.CloseTime,
			WonTime:          item.WonTime,
			LostTime:         item.LostTime,
		}

		if len(results) >= limit {
			break
		}
		results = append(results, summary)
	}

	return results
}

// filterAndSummarizeListedDeals applies client-side filters to listed deals.
func (c *Client) filterAndSummarizeListedDeals(data json.RawMessage, daysBack int, params GetDealsParams, limit int) []DealSummary {
	var rawDeals []json.RawMessage
	if err := json.Unmarshal(data, &rawDeals); err != nil {
		return nil
	}

	cutoff := time.Now().AddDate(0, 0, -daysBack)
	var results []DealSummary

	for _, raw := range rawDeals {
		var deal RawDeal
		if err := json.Unmarshal(raw, &deal); err != nil {
			continue
		}

		// Filter by last_activity_date cutoff
		if deal.LastActivityDate != "" {
			t, err := time.Parse("2006-01-02", deal.LastActivityDate)
			if err == nil && t.Before(cutoff) {
				continue
			}
		}

		// Apply min/max value filters
		if params.MinValue != nil && deal.Value < *params.MinValue {
			continue
		}
		if params.MaxValue != nil && deal.Value > *params.MaxValue {
			continue
		}

		// Extract nested names
		stageName := deal.StageName
		if stageName == "" {
			stageName = "Unknown"
		}
		pipelineName := deal.PipelineName
		if pipelineName == "" {
			pipelineName = "Unknown"
		}
		ownerName := deal.OwnerName
		if ownerName == "" {
			ownerName = "Unknown"
		}

		var orgName *string
		if deal.OrgID != nil && deal.OrgID.GetName() != "" {
			n := deal.OrgID.GetName()
			orgName = &n
		}
		var personName *string
		if deal.PersonID != nil && deal.PersonID.GetName() != "" {
			n := deal.PersonID.GetName()
			personName = &n
		}

		// Extract booking details from custom fields
		var bookingDetails *string
		var rawMap map[string]interface{}
		if json.Unmarshal(raw, &rawMap) == nil {
			if v, ok := rawMap[c.bookingFieldKey]; ok {
				if s, ok := v.(string); ok && s != "" {
					bookingDetails = &s
				}
			}
		}

		summary := DealSummary{
			ID:               deal.ID,
			Title:            deal.Title,
			Value:            deal.Value,
			Currency:         deal.Currency,
			Status:           deal.Status,
			StageName:        stageName,
			PipelineName:     pipelineName,
			OwnerName:        ownerName,
			OrganizationName: orgName,
			PersonName:       personName,
			AddTime:          deal.AddTime,
			LastActivityDate: deal.LastActivityDate,
			CloseTime:        deal.CloseTime,
			WonTime:          deal.WonTime,
			LostTime:         deal.LostTime,
			NotesCount:       deal.NotesCount,
			BookingDetails:   bookingDetails,
		}

		if len(results) >= limit {
			break
		}
		results = append(results, summary)
	}

	return results
}

// GetDeal fetches a specific deal by ID.
func (c *Client) GetDeal(ctx context.Context, dealID int) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("deals/%d", dealID)
	return c.getRaw(ctx, endpoint, nil)
}

// SearchDeals searches deals by term.
func (c *Client) SearchDeals(ctx context.Context, term string) (json.RawMessage, error) {
	return c.getRaw(ctx, "deals/search", map[string]string{"term": term})
}
