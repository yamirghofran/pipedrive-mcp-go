package pipedrive

import (
	"context"
	"encoding/json"
)

// SearchAll performs a cross-entity search across all item types.
func (c *Client) SearchAll(ctx context.Context, term string, itemTypes string) (json.RawMessage, error) {
	params := map[string]string{
		"term": term,
	}
	if itemTypes != "" {
		params["item_type"] = itemTypes
	}
	return c.getRaw(ctx, "search", params)
}
