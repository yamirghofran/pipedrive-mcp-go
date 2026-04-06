package pipedrive

import (
	"context"
	"encoding/json"
)

// User represents a Pipedrive user/owner.
type User struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	ActiveFlag bool   `json:"active_flag"`
	RoleName   string `json:"role_name"`
}

// GetUsers fetches all users from Pipedrive.
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	data, err := c.getRaw(ctx, "users", nil)
	if err != nil {
		return nil, err
	}

	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}
	return users, nil
}
