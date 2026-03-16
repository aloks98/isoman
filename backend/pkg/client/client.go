package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is an HTTP client for the ISOMan API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// WithUserAgent sets the User-Agent header sent with every request.
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua }
}

// NewClient creates a new ISOMan API client.
// baseURL is the root URL of the ISOMan server (e.g. "http://localhost:8080").
func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
		userAgent:  "isoman-go-client",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// apiResponse is the raw JSON envelope returned by all ISOMan API endpoints.
type apiResponse struct {
	Success bool             `json:"success"`
	Data    json.RawMessage  `json:"data,omitempty"`
	Error   *apiResponseError `json:"error,omitempty"`
	Message string           `json:"message,omitempty"`
}

type apiResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// do executes an HTTP request and returns the raw response.
func (c *Client) do(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	reqURL := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("isoman: build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("isoman: %s %s: %w", method, path, err)
	}
	return resp, nil
}

// doJSON executes a request, decodes the API envelope, and unmarshals data into dest.
// If dest is nil the data field is ignored (useful for DELETE).
func (c *Client) doJSON(ctx context.Context, method, path string, body io.Reader, dest any) error {
	resp, err := c.do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var envelope apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("isoman: decode response: %w", err)
	}

	if !envelope.Success {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if envelope.Error != nil {
			apiErr.Code = envelope.Error.Code
			apiErr.Message = envelope.Error.Message
			apiErr.Details = envelope.Error.Details
		}
		return apiErr
	}

	if dest != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, dest); err != nil {
			return fmt.Errorf("isoman: unmarshal data: %w", err)
		}
	}
	return nil
}

// encodeBody marshals v to a JSON reader.
func encodeBody(v any) (io.Reader, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("isoman: encode request: %w", err)
	}
	return strings.NewReader(string(b)), nil
}

// ListISOs returns a paginated list of ISOs.
// Pass nil for default options (page 1, page_size 10, sorted by created_at desc).
func (c *Client) ListISOs(ctx context.Context, opts *ListISOsOptions) (*ListISOsResponse, error) {
	path := "/api/isos"
	if opts != nil {
		q := url.Values{}
		if opts.Page > 0 {
			q.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.PageSize > 0 {
			q.Set("page_size", strconv.Itoa(opts.PageSize))
		}
		if opts.SortBy != "" {
			q.Set("sort_by", opts.SortBy)
		}
		if opts.SortDir != "" {
			q.Set("sort_dir", opts.SortDir)
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var result ListISOsResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetISO returns a single ISO by ID.
func (c *Client) GetISO(ctx context.Context, id string) (*ISO, error) {
	var iso ISO
	if err := c.doJSON(ctx, http.MethodGet, "/api/isos/"+id, nil, &iso); err != nil {
		return nil, err
	}
	return &iso, nil
}

// CreateISO queues a new ISO download and returns the created ISO.
func (c *Client) CreateISO(ctx context.Context, req CreateISORequest) (*ISO, error) {
	body, err := encodeBody(req)
	if err != nil {
		return nil, err
	}
	var iso ISO
	if err := c.doJSON(ctx, http.MethodPost, "/api/isos", body, &iso); err != nil {
		return nil, err
	}
	return &iso, nil
}

// UpdateISO updates an existing ISO and returns the updated ISO.
func (c *Client) UpdateISO(ctx context.Context, id string, req UpdateISORequest) (*ISO, error) {
	body, err := encodeBody(req)
	if err != nil {
		return nil, err
	}
	var iso ISO
	if err := c.doJSON(ctx, http.MethodPut, "/api/isos/"+id, body, &iso); err != nil {
		return nil, err
	}
	return &iso, nil
}

// DeleteISO deletes an ISO by ID, removing the file and database record.
func (c *Client) DeleteISO(ctx context.Context, id string) error {
	return c.doJSON(ctx, http.MethodDelete, "/api/isos/"+id, nil, nil)
}

// RetryISO retries a failed ISO download and returns the updated ISO.
func (c *Client) RetryISO(ctx context.Context, id string) (*ISO, error) {
	var iso ISO
	if err := c.doJSON(ctx, http.MethodPost, "/api/isos/"+id+"/retry", nil, &iso); err != nil {
		return nil, err
	}
	return &iso, nil
}

// GetStats returns aggregated statistics.
func (c *Client) GetStats(ctx context.Context) (*Stats, error) {
	var stats Stats
	if err := c.doJSON(ctx, http.MethodGet, "/api/stats", nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetDownloadTrends returns download trend data over time.
// Pass nil for default options (daily period, 30 days).
func (c *Client) GetDownloadTrends(ctx context.Context, opts *DownloadTrendsOptions) (*DownloadTrends, error) {
	path := "/api/stats/trends"
	if opts != nil {
		q := url.Values{}
		if opts.Period != "" {
			q.Set("period", opts.Period)
		}
		if opts.Days > 0 {
			q.Set("days", strconv.Itoa(opts.Days))
		}
		if encoded := q.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	var trends DownloadTrends
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &trends); err != nil {
		return nil, err
	}
	return &trends, nil
}

// Health checks whether the ISOMan server is healthy.
// Returns nil if healthy, or an error otherwise.
func (c *Client) Health(ctx context.Context) error {
	return c.doJSON(ctx, http.MethodGet, "/health", nil, nil)
}

// DownloadFile downloads a file from the /images/ endpoint.
// filePath is the path relative to /images/ (e.g. "alpine/3.19.1/x86_64/alpine-3.19.1-x86_64.iso").
// The caller is responsible for closing the returned ReadCloser.
func (c *Client) DownloadFile(ctx context.Context, filePath string) (io.ReadCloser, error) {
	path := "/images/" + strings.TrimLeft(filePath, "/")
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Code:       "DOWNLOAD_FAILED",
			Message:    fmt.Sprintf("unexpected status %d for %s", resp.StatusCode, path),
		}
	}
	return resp.Body, nil
}
