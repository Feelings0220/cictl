// Package jenkins is a minimal read-only Jenkins REST client.
package jenkins

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Feelings0220/cictl/internal/config"
)

type Client struct {
	base  *url.URL
	user  string
	token string
	http  *http.Client
}

func New(c config.Credentials) (*Client, error) {
	if c.URL == "" {
		return nil, fmt.Errorf("jenkins url is empty")
	}
	u, err := url.Parse(strings.TrimRight(c.URL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse jenkins url: %w", err)
	}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: c.Insecure}} //nolint:gosec // opt-in via config
	return &Client{
		base:  u,
		user:  c.Username,
		token: c.Token,
		http:  &http.Client{Transport: tr, Timeout: 30 * time.Second},
	}, nil
}

// GET performs an authenticated GET. path begins with "/".
// If into is non-nil and response is JSON, it is unmarshalled into it.
// Raw response body is always returned for callers that want bytes.
func (c *Client) GET(ctx context.Context, path string, into any) ([]byte, error) {
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must begin with /")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base.String()+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(c.user, c.token)
	req.Header.Set("Accept", "application/json,application/xml,text/plain;q=0.9,*/*;q=0.5")
	resp, err := c.http.Do(req)
	if err != nil {
		// Strip credentials from any url errors.
		return nil, fmt.Errorf("jenkins request failed: %w", scrubURLError(err, c.token))
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("jenkins returned %d for %s", resp.StatusCode, path)
	}
	if into != nil {
		if err := json.Unmarshal(body, into); err != nil {
			return body, fmt.Errorf("decode json from %s: %w", path, err)
		}
	}
	return body, nil
}

// scrubURLError removes the token from net/url errors that embed the request URL.
func scrubURLError(err error, token string) error {
	if token == "" {
		return err
	}
	return fmt.Errorf("%s", strings.ReplaceAll(err.Error(), token, "***"))
}
