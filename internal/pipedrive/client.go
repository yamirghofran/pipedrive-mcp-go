// Package pipedrive provides a Go client for the Pipedrive API v1.
package pipedrive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// UserAgent is the User-Agent header sent with all Pipedrive API requests.
const UserAgent = "pipedrive-mcp-server/2.0.0 (go)"

// DefaultBookingFieldKey is the default custom field key for booking details.
const DefaultBookingFieldKey = "8f4b27fbd9dfc70d2296f23ce76987051ad7324e"

// Client is a Pipedrive API client with rate limiting.
// It supports both v1 and v2 API endpoints.
type Client struct {
	httpClient      *http.Client
	baseURL         string // v1 API base URL
	baseURLV2       string // v2 API base URL
	apiToken        string
	limiter         *rate.Limiter
	bookingFieldKey string
}

// NewClient creates a new Pipedrive API client.
func NewClient(domain, apiToken string, minInterval time.Duration, maxConcurrent int, bookingFieldKey string) *Client {
	// rate.Limiter: r = 1/interval, b = maxConcurrent
	ratePerSecond := float64(1) / minInterval.Seconds()
	if ratePerSecond <= 0 {
		ratePerSecond = 4 // fallback: 250ms interval
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:         fmt.Sprintf("https://%s/api/v1", domain),
		baseURLV2:       fmt.Sprintf("https://%s/api/v2", domain),
		apiToken:        apiToken,
		limiter:         rate.NewLimiter(rate.Limit(ratePerSecond), maxConcurrent),
		bookingFieldKey: bookingFieldKey,
	}
}

// PipedriveResponse is the common response envelope for single-item responses.
type PipedriveResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

// PipedrivePaginatedResponse is the common response envelope for paginated responses.
type PipedrivePaginatedResponse struct {
	Success        bool            `json:"success"`
	Data           json.RawMessage `json:"data"`
	AdditionalData *PaginationData `json:"additional_data,omitempty"`
}

// PaginationData holds pagination metadata.
type PaginationData struct {
	Pagination struct {
		Start                 int  `json:"start"`
		Limit                 int  `json:"limit"`
		MoreItemsInCollection bool `json:"more_items_in_collection"`
	} `json:"pagination"`
}

// PipedriveError represents an error from the Pipedrive API.
type PipedriveError struct {
	StatusCode int
	Message    string // sanitized, safe to return to client
	Detail     string // full detail, for logging only
}

func (e *PipedriveError) Error() string {
	return e.Message
}

// sanitizeURL removes the api_token query parameter from a URL string.
func sanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	q.Del("api_token")
	u.RawQuery = q.Encode()
	return u.String()
}

// tokenPattern matches api_token query parameters in URLs.
var tokenPattern = regexp.MustCompile(`api_token=[^&\s]+`)

// sanitizeString removes api_token from any URL found in a string.
func sanitizeString(s string) string {
	return tokenPattern.ReplaceAllString(s, "api_token=[REDACTED]")
}

// get performs a GET request to the Pipedrive API.
func (c *Client) get(ctx context.Context, endpoint string, params map[string]string) (*http.Response, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, &PipedriveError{
			StatusCode: http.StatusTooManyRequests,
			Message:    "rate limit exceeded, please try again",
			Detail:     err.Error(),
		}
	}

	reqURL := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("api_token", c.apiToken)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	slog.Debug("Pipedrive API request", "method", "GET", "endpoint", endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &PipedriveError{
			StatusCode: 0,
			Message:    "failed to connect to Pipedrive API",
			Detail:     sanitizeString(err.Error()),
		}
	}

	return resp, nil
}

// getRaw fetches and returns the raw JSON data from a Pipedrive API endpoint.
// Returns the raw "data" field as json.RawMessage.
func (c *Client) getRaw(ctx context.Context, endpoint string, params map[string]string) (json.RawMessage, error) {
	resp, err := c.get(ctx, endpoint, params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to read Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if resp.StatusCode >= 400 {
		slog.Error("Pipedrive API error", "status", resp.StatusCode, "body", sanitizeString(string(body)))
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Pipedrive API returned status %d", resp.StatusCode),
			Detail:     sanitizeString(string(body)),
		}
	}

	var result PipedriveResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to parse Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if !result.Success {
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "Pipedrive API returned unsuccessful response",
			Detail:     sanitizeString(string(body)),
		}
	}

	return result.Data, nil
}

// getList fetches paginated data from a Pipedrive API endpoint.
func (c *Client) getList(ctx context.Context, endpoint string, params map[string]string) (json.RawMessage, *PaginationData, error) {
	resp, err := c.get(ctx, endpoint, params)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to read Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if resp.StatusCode >= 400 {
		slog.Error("Pipedrive API error", "status", resp.StatusCode, "body", sanitizeString(string(body)))
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Pipedrive API returned status %d", resp.StatusCode),
			Detail:     sanitizeString(string(body)),
		}
	}

	var result PipedrivePaginatedResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to parse Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if !result.Success {
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "Pipedrive API returned unsuccessful response",
			Detail:     sanitizeString(string(body)),
		}
	}

	return result.Data, result.AdditionalData, nil
}

// PipedriveV2PaginatedResponse is the response envelope for v2 cursor-based pagination.
type PipedriveV2PaginatedResponse struct {
	Success        bool            `json:"success"`
	Data           json.RawMessage `json:"data"`
	AdditionalData *struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"additional_data,omitempty"`
}

// PaginationDataV2 holds cursor-based pagination metadata for v2 API.
type PaginationDataV2 struct {
	NextCursor *string
}

// getListV2 fetches paginated data from a v2 API endpoint (cursor-based pagination).
func (c *Client) getListV2(ctx context.Context, endpoint string, params map[string]string) (json.RawMessage, *PaginationDataV2, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, nil, &PipedriveError{
			StatusCode: http.StatusTooManyRequests,
			Message:    "rate limit exceeded, please try again",
			Detail:     err.Error(),
		}
	}

	reqURL := fmt.Sprintf("%s/%s", c.baseURLV2, cleanEndpoint(endpoint))
	u, err := url.Parse(reqURL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid URL: %w", err)
	}

	q := u.Query()
	q.Set("api_token", c.apiToken)
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	slog.Debug("Pipedrive v2 API request", "method", "GET", "endpoint", endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, &PipedriveError{
			StatusCode: 0,
			Message:    "failed to connect to Pipedrive API",
			Detail:     sanitizeString(err.Error()),
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to read Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if resp.StatusCode >= 400 {
		slog.Error("Pipedrive v2 API error", "status", resp.StatusCode, "body", sanitizeString(string(body)))
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Pipedrive API returned status %d", resp.StatusCode),
			Detail:     sanitizeString(string(body)),
		}
	}

	var result PipedriveV2PaginatedResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to parse Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if !result.Success {
		return nil, nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "Pipedrive API returned unsuccessful response",
			Detail:     sanitizeString(string(body)),
		}
	}

	var pagination *PaginationDataV2
	if result.AdditionalData != nil {
		pagination = &PaginationDataV2{
			NextCursor: result.AdditionalData.NextCursor,
		}
	}

	return result.Data, pagination, nil
}

// getRawV2 fetches a single item from a v2 API endpoint.
func (c *Client) getRawV2(ctx context.Context, endpoint string, params map[string]string) (json.RawMessage, error) {
	data, _, err := c.getListV2(ctx, endpoint, params)
	return data, err
}

// doWrite performs a write request (POST/PATCH/PUT/DELETE) to the specified base URL.
func (c *Client) doWrite(ctx context.Context, method, baseURL, endpoint string, body interface{}) (json.RawMessage, error) {
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, &PipedriveError{
			StatusCode: http.StatusTooManyRequests,
			Message:    "rate limit exceeded, please try again",
			Detail:     err.Error(),
		}
	}

	reqURL := fmt.Sprintf("%s/%s", baseURL, cleanEndpoint(endpoint))

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Add API token as query parameter
	q := req.URL.Query()
	q.Set("api_token", c.apiToken)
	req.URL.RawQuery = q.Encode()

	slog.Debug("Pipedrive API request", "method", method, "endpoint", endpoint)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &PipedriveError{
			StatusCode: 0,
			Message:    "failed to connect to Pipedrive API",
			Detail:     sanitizeString(err.Error()),
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to read Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if resp.StatusCode >= 400 {
		slog.Error("Pipedrive API error", "method", method, "status", resp.StatusCode, "body", sanitizeString(string(respBody)))
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Pipedrive API returned status %d", resp.StatusCode),
			Detail:     sanitizeString(string(respBody)),
		}
	}

	// For DELETE requests, the response may not have a standard envelope
	if method == http.MethodDelete {
		var result struct {
			Success bool            `json:"success"`
			Data    json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(respBody, &result); err != nil {
			// If we can't parse, just return success
			return json.RawMessage(`{"success": true}`), nil
		}
		return result.Data, nil
	}

	var result PipedriveResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "failed to parse Pipedrive API response",
			Detail:     err.Error(),
		}
	}

	if !result.Success {
		return nil, &PipedriveError{
			StatusCode: resp.StatusCode,
			Message:    "Pipedrive API returned unsuccessful response",
			Detail:     sanitizeString(string(respBody)),
		}
	}

	return result.Data, nil
}

// postV1 performs a POST request to the v1 API.
func (c *Client) postV1(ctx context.Context, endpoint string, body interface{}) (json.RawMessage, error) {
	return c.doWrite(ctx, http.MethodPost, c.baseURL, endpoint, body)
}

// putV1 performs a PUT request to the v1 API.
func (c *Client) putV1(ctx context.Context, endpoint string, body interface{}) (json.RawMessage, error) {
	return c.doWrite(ctx, http.MethodPut, c.baseURL, endpoint, body)
}

// deleteV1 performs a DELETE request to the v1 API.
func (c *Client) deleteV1(ctx context.Context, endpoint string) (json.RawMessage, error) {
	return c.doWrite(ctx, http.MethodDelete, c.baseURL, endpoint, nil)
}

// postV2 performs a POST request to the v2 API.
func (c *Client) postV2(ctx context.Context, endpoint string, body interface{}) (json.RawMessage, error) {
	return c.doWrite(ctx, http.MethodPost, c.baseURLV2, endpoint, body)
}

// patchV2 performs a PATCH request to the v2 API.
func (c *Client) patchV2(ctx context.Context, endpoint string, body interface{}) (json.RawMessage, error) {
	return c.doWrite(ctx, http.MethodPatch, c.baseURLV2, endpoint, body)
}

// deleteV2 performs a DELETE request to the v2 API.
func (c *Client) deleteV2(ctx context.Context, endpoint string) (json.RawMessage, error) {
	return c.doWrite(ctx, http.MethodDelete, c.baseURLV2, endpoint, nil)
}

// BookingFieldKey returns the configured booking field key.
func (c *Client) BookingFieldKey() string {
	return c.bookingFieldKey
}

// extractStringValue extracts a string value from a raw JSON message by key.
func extractStringValue(data json.RawMessage, key string) *string {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return nil
	}
	return &s
}

// SanitizeErrorMessage creates a sanitized error message safe to return to clients.
func SanitizeErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	// Strip api_token from any URLs in the message
	return sanitizeString(msg)
}

// buildParams creates a parameter map from non-nil/non-zero values.
func buildParams(entries ...struct{ k, v string }) map[string]string {
	params := make(map[string]string)
	for _, e := range entries {
		if e.v != "" {
			params[e.k] = e.v
		}
	}
	if len(params) == 0 {
		return nil
	}
	return params
}

// ptrToStr safely dereferences a string pointer, returning "" for nil.
func ptrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// intPtr returns a pointer to the given int.
func intPtr(n int) *int {
	return &n
}

// intToStr converts an int to string, returning "" for nil pointer.
func intToStr(n *int) string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("%d", *n)
}

// floatToStr converts a float64 to string, returning "" for nil pointer.
func floatToStr(f *float64) string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("%g", *f)
}

// ensure trailing slash consistency
func cleanEndpoint(ep string) string {
	return strings.TrimPrefix(ep, "/")
}
