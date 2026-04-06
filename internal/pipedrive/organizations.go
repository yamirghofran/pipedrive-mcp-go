package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetOrganizations fetches all organizations from Pipedrive.
func (c *Client) GetOrganizations(ctx context.Context) (json.RawMessage, error) {
	return c.getRaw(ctx, "organizations", nil)
}

// GetOrganization fetches a specific organization by ID.
func (c *Client) GetOrganization(ctx context.Context, orgID int) (json.RawMessage, error) {
	return c.getRaw(ctx, fmt.Sprintf("organizations/%d", orgID), nil)
}

// SearchOrganizations searches organizations by term.
func (c *Client) SearchOrganizations(ctx context.Context, term string) (json.RawMessage, error) {
	return c.getRaw(ctx, "organizations/search", map[string]string{"term": term})
}
