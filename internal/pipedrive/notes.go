package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
)

// Note represents a Pipedrive note.
type Note struct {
	ID               int    `json:"id"`
	Content          string `json:"content"`
	DealID           int    `json:"deal_id"`
	PersonID         int    `json:"person_id"`
	OrgID            int    `json:"org_id"`
	AddTime          string `json:"add_time"`
	UpdateTime       string `json:"update_time"`
	PinnedToDeal     bool   `json:"pinned_to_deal_flag"`
	PinnedToPerson   bool   `json:"pinned_to_person_flag"`
	PinnedToOrg      bool   `json:"pinned_to_organization_flag"`
	LastUpdateUserID int    `json:"last_update_user_id"`
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
