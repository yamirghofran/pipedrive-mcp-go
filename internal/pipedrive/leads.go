package pipedrive

import (
	"context"
	"encoding/json"
)

// SearchLeads searches leads by term.
func (c *Client) SearchLeads(ctx context.Context, term string) (json.RawMessage, error) {
	return c.getRaw(ctx, "leads/search", map[string]string{"term": term})
}
