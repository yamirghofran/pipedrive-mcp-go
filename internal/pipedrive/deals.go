package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// --- Data types for deal-related responses ---

// RawDealV2 represents a deal from the Pipedrive v2 list deals endpoint.
// The v2 API returns flat IDs instead of nested objects.
type RawDealV2 struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	OwnerID    int     `json:"owner_id"`
	PersonID   int     `json:"person_id"`
	OrgID      int     `json:"org_id"`
	PipelineID int     `json:"pipeline_id"`
	StageID    int     `json:"stage_id"`
	Value      float64 `json:"value"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
	AddTime    string  `json:"add_time"`
	UpdateTime string  `json:"update_time"`
	CloseTime  string  `json:"close_time"`
	WonTime    string  `json:"won_time"`
	LostTime   string  `json:"lost_time"`
	NotesCount int     `json:"notes_count"`
	// Custom fields stored as raw JSON object (40-char hash keys)
	CustomFields json.RawMessage `json:"custom_fields"`
}

// NestedIDName represents a nested object with id and name fields (v1 API format).
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
// Uses 'any' for ID fields to flexibly accept integers, strings, or null from MCP clients.
type GetDealsParams struct {
	SearchTitle any `json:"searchTitle,omitempty" jsonschema:"Search deals by title (partial matches)"`
	DaysBack    any `json:"daysBack,omitempty" jsonschema:"Days back to fetch based on update time. Default: 365"`
	OwnerID     any `json:"ownerId,omitempty" jsonschema:"Filter by owner/user ID"`
	StageID     any `json:"stageId,omitempty" jsonschema:"Filter by stage ID"`
	Status      any `json:"status,omitempty" jsonschema:"Deal status: open, won, lost, or deleted. Default: returns all non-deleted deals"`
	PipelineID  any `json:"pipelineId,omitempty" jsonschema:"Filter by pipeline ID"`
	MinValue    any `json:"minValue,omitempty" jsonschema:"Minimum deal value"`
	MaxValue    any `json:"maxValue,omitempty" jsonschema:"Maximum deal value"`
	Limit       any `json:"limit,omitempty" jsonschema:"Maximum deals to return (capped at 500). Default: 500"`
}

// GetDealsResult is the response for the get-deals tool.
type GetDealsResult struct {
	Summary        string                 `json:"summary"`
	FiltersApplied map[string]interface{} `json:"filters_applied"`
	TotalFound     int                    `json:"total_found"`
	Deals          []DealSummary          `json:"deals"`
}

// dealLookups holds name lookup maps for enriching v2 deal data.
type dealLookups struct {
	stageNameMap    map[int]string // stage_id  -> stage name
	pipelineNameMap map[int]string // pipeline_id -> pipeline name
	ownerNameMap    map[int]string // owner_id  -> owner name
}

// fetchDealLookups concurrently fetches pipelines, stages, and users to build
// name lookup maps used for enriching v2 deal responses.
func (c *Client) fetchDealLookups(ctx context.Context) *dealLookups {
	var (
		stageNameMap    = make(map[int]string)
		pipelineNameMap = make(map[int]string)
		ownerNameMap    = make(map[int]string)
		wg              sync.WaitGroup
	)

	// Fetch pipelines (v1 API)
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := c.getRaw(ctx, "pipelines", nil)
		if err != nil {
			return
		}
		var pipelines []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &pipelines); err != nil {
			return
		}
		for _, p := range pipelines {
			pipelineNameMap[p.ID] = p.Name
		}
	}()

	// Fetch stages (v2 API — returns all stages in one call)
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, _, err := c.getListV2(ctx, "stages", map[string]string{"limit": "500"})
		if err != nil {
			return
		}
		var stages []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &stages); err != nil {
			return
		}
		for _, s := range stages {
			stageNameMap[s.ID] = s.Name
		}
	}()

	// Fetch users (v1 API)
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := c.getRaw(ctx, "users", nil)
		if err != nil {
			return
		}
		var users []User
		if err := json.Unmarshal(data, &users); err != nil {
			return
		}
		for _, u := range users {
			ownerNameMap[u.ID] = u.Name
		}
	}()

	wg.Wait()

	return &dealLookups{
		stageNameMap:    stageNameMap,
		pipelineNameMap: pipelineNameMap,
		ownerNameMap:    ownerNameMap,
	}
}

// GetDeals fetches deals with flexible filtering using the v2 API.
func (c *Client) GetDeals(ctx context.Context, params GetDealsParams) (*GetDealsResult, error) {
	// Apply defaults
	daysBack := 365
	if v, err := ParseIntField(params.DaysBack); err == nil && v != nil {
		daysBack = *v
	}
	// Default to empty string so the API returns all non-deleted deals.
	// The v2 API docs state: "If omitted, all not deleted deals are returned."
	status := ""
	if v, err := ParseStringField(params.Status); err == nil && v != nil && *v != "" {
		status = *v
	}
	limit := 500
	if v, err := ParseIntField(params.Limit); err == nil && v != nil && *v < 500 {
		limit = *v
	}

	var dealSummaries []DealSummary
	var filtersApplied = make(map[string]interface{})

	if searchTitle, _ := ParseStringField(params.SearchTitle); searchTitle != nil && *searchTitle != "" {
		// Search path (still uses v1 search endpoint)
		filtersApplied["searchTitle"] = *searchTitle
		searchResults, err := c.searchDealsRaw(ctx, *searchTitle)
		if err != nil {
			return nil, err
		}

		dealSummaries = c.filterAndSummarizeSearchResults(searchResults, params, limit)
	} else {
		// List path — use v2 API (GET /api/v2/deals)
		apiParams := map[string]string{
			"sort_by":        "update_time",
			"sort_direction": "desc",
			"limit":          fmt.Sprintf("%d", limit),
		}

		// Apply status filter server-side
		if status != "" {
			apiParams["status"] = status
			filtersApplied["status"] = status
		}

		// Apply owner filter server-side using v2 parameter name
		if v, err := ParseIntField(params.OwnerID); err == nil && v != nil {
			apiParams["owner_id"] = fmt.Sprintf("%d", *v)
			filtersApplied["ownerId"] = *v
		}
		if v, err := ParseIntField(params.StageID); err == nil && v != nil {
			apiParams["stage_id"] = fmt.Sprintf("%d", *v)
			filtersApplied["stageId"] = *v
		}
		if v, err := ParseIntField(params.PipelineID); err == nil && v != nil {
			apiParams["pipeline_id"] = fmt.Sprintf("%d", *v)
			filtersApplied["pipelineId"] = *v
		}

		// Apply daysBack filter server-side using updated_since parameter
		if daysBack > 0 {
			cutoff := time.Now().AddDate(0, 0, -daysBack).UTC().Format(time.RFC3339)
			apiParams["updated_since"] = cutoff
			filtersApplied["daysBack"] = daysBack
		}

		// Request notes_count in the response
		apiParams["include_fields"] = "notes_count"

		data, _, err := c.getListV2(ctx, "deals", apiParams)
		if err != nil {
			return nil, err
		}

		// Fetch name lookups concurrently for enriching deal data
		lookups := c.fetchDealLookups(ctx)

		dealSummaries = c.filterAndSummarizeListedDealsV2(data, params, lookups, limit)
	}

	// Track additional filters
	if v, err := ParseFloatField(params.MinValue); err == nil && v != nil {
		filtersApplied["minValue"] = *v
	}
	if v, err := ParseFloatField(params.MaxValue); err == nil && v != nil {
		filtersApplied["maxValue"] = *v
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

	// Pre-parse flexible params for use in the loop
	ownerID, _ := ParseIntField(params.OwnerID)
	statusVal, _ := ParseStringField(params.Status)
	stageID, _ := ParseIntField(params.StageID)
	pipelineID, _ := ParseIntField(params.PipelineID)
	minValue, _ := ParseFloatField(params.MinValue)
	maxValue, _ := ParseFloatField(params.MaxValue)

	for _, item := range items {
		// Apply owner filter
		if ownerID != nil {
			uid, _ := item.UserID.Int64()
			if int(uid) != *ownerID {
				continue
			}
		}

		// Apply status filter
		if statusVal != nil && *statusVal != "" && item.Status != *statusVal {
			continue
		}

		// Apply stage filter
		if stageID != nil && item.StageID != *stageID {
			continue
		}

		// Apply pipeline filter
		if pipelineID != nil && item.PipelineID != *pipelineID {
			continue
		}

		// Apply min/max value filters
		if minValue != nil && item.Value < *minValue {
			continue
		}
		if maxValue != nil && item.Value > *maxValue {
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

// filterAndSummarizeListedDealsV2 applies client-side filters to v2 listed deals
// and enriches them with name lookups.
func (c *Client) filterAndSummarizeListedDealsV2(data json.RawMessage, params GetDealsParams, lookups *dealLookups, limit int) []DealSummary {
	var deals []RawDealV2
	if err := json.Unmarshal(data, &deals); err != nil {
		return nil
	}

	// Pre-parse flexible params for use in the loop
	minValue, _ := ParseFloatField(params.MinValue)
	maxValue, _ := ParseFloatField(params.MaxValue)

	var results []DealSummary

	for _, deal := range deals {
		// Apply min/max value filters (client-side)
		if minValue != nil && deal.Value < *minValue {
			continue
		}
		if maxValue != nil && deal.Value > *maxValue {
			continue
		}

		// Resolve names from lookups
		stageName := "Unknown"
		if name, ok := lookups.stageNameMap[deal.StageID]; ok && name != "" {
			stageName = name
		}
		pipelineName := "Unknown"
		if name, ok := lookups.pipelineNameMap[deal.PipelineID]; ok && name != "" {
			pipelineName = name
		}
		ownerName := "Unknown"
		if name, ok := lookups.ownerNameMap[deal.OwnerID]; ok && name != "" {
			ownerName = name
		}

		// Extract booking details from custom fields
		var bookingDetails *string
		if deal.CustomFields != nil {
			var cfMap map[string]interface{}
			if json.Unmarshal(deal.CustomFields, &cfMap) == nil {
				if v, ok := cfMap[c.bookingFieldKey]; ok {
					if s, ok := v.(string); ok && s != "" {
						bookingDetails = &s
					}
				}
			}
		}

		summary := DealSummary{
			ID:             deal.ID,
			Title:          deal.Title,
			Value:          deal.Value,
			Currency:       deal.Currency,
			Status:         deal.Status,
			StageName:      stageName,
			PipelineName:   pipelineName,
			OwnerName:      ownerName,
			AddTime:        deal.AddTime,
			CloseTime:      deal.CloseTime,
			WonTime:        deal.WonTime,
			LostTime:       deal.LostTime,
			NotesCount:     deal.NotesCount,
			BookingDetails: bookingDetails,
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
