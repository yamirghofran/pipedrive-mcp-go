package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetPersons fetches all persons from Pipedrive.
func (c *Client) GetPersons(ctx context.Context) (json.RawMessage, error) {
	return c.getRaw(ctx, "persons", nil)
}

// GetPerson fetches a specific person by ID.
func (c *Client) GetPerson(ctx context.Context, personID int) (json.RawMessage, error) {
	return c.getRaw(ctx, fmt.Sprintf("persons/%d", personID), nil)
}

// SearchPersons searches persons by term.
func (c *Client) SearchPersons(ctx context.Context, term string) (json.RawMessage, error) {
	return c.getRaw(ctx, "persons/search", map[string]string{"term": term})
}

// Person represents a Pipedrive person (contact).
type Person struct {
	ID    int             `json:"id"`
	Name  string          `json:"name"`
	Email []EmailField    `json:"email"`
	Phone []PhoneField    `json:"phone"`
	OrgID *NestedIDName   `json:"org_id"`
	Raw   json.RawMessage `json:"-"`
}

// EmailField represents an email field in Pipedrive.
type EmailField struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// PhoneField represents a phone field in Pipedrive.
type PhoneField struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
