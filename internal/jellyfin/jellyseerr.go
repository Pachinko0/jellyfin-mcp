package jellyfin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// JellyseerrClient is the HTTP client used by the read-only watchlist tool.
type JellyseerrClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// JellyseerrHTTPError retains only the status code so response bodies cannot
// accidentally leak credentials or sensitive server details through errors.
type JellyseerrHTTPError struct{ StatusCode int }

func (e *JellyseerrHTTPError) Error() string {
	if e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden {
		return fmt.Sprintf("Jellyseerr authentication failed (HTTP %d)", e.StatusCode)
	}
	return fmt.Sprintf("Jellyseerr API request failed (HTTP %d)", e.StatusCode)
}

func IsJellyseerrAuthError(err error) bool {
	e, ok := err.(*JellyseerrHTTPError)
	return ok && (e.StatusCode == http.StatusUnauthorized || e.StatusCode == http.StatusForbidden)
}

// NewJellyseerrClient reads connection settings without logging them.
func NewJellyseerrClient() (*JellyseerrClient, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("JELLYSEERR_URL")), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("JELLYSEERR_URL is not configured")
	}
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("JELLYSEERR_URL is invalid")
	}
	apiKey := strings.TrimSpace(os.Getenv("JELLYSEERR_API_KEY"))
	if apiKey == "" {
		return nil, fmt.Errorf("JELLYSEERR_API_KEY is not configured")
	}
	return &JellyseerrClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *JellyseerrClient) Get(ctx context.Context, endpoint string, params url.Values, jellyseerrUserID string, dest any) error {
	u, err := url.JoinPath(c.baseURL, endpoint)
	if err != nil {
		return fmt.Errorf("building Jellyseerr URL")
	}
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, bytes.NewReader(nil))
	if err != nil {
		return fmt.Errorf("creating Jellyseerr request")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	if jellyseerrUserID != "" {
		req.Header.Set("X-Api-User", jellyseerrUserID)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("connecting to Jellyseerr: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &JellyseerrHTTPError{StatusCode: resp.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxResponseBodyBytes))
	if err != nil {
		return fmt.Errorf("reading Jellyseerr response")
	}
	if dest != nil && len(body) > 0 {
		if err := json.Unmarshal(body, dest); err != nil {
			return fmt.Errorf("decoding Jellyseerr response")
		}
	}
	return nil
}
