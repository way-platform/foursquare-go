package foursquare

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://places-api.foursquare.com"
	defaultTimeout = 30 * time.Second

	// apiVersion is the pinned X-Places-Api-Version header value.
	apiVersion = "2025-06-17"
)

// Client is a client for the Foursquare Places API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	transport  http.RoundTripper
	timeout    time.Duration
	retryCount int
	retrySleep func(ctx context.Context, d time.Duration) bool
}

// Option configures a [Client].
type Option func(*Client)

// WithAPIKey sets the Foursquare API key used as a Bearer token on every request.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = key }
}

// WithBaseURL overrides the default API base URL. Useful for testing.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient sets a base *http.Client whose Transport is used as the
// innermost transport in the chain.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTransport sets the base [http.RoundTripper], replacing [http.DefaultTransport].
// Use this to inject observability (metrics, tracing) without adding any
// vendor-specific code to this SDK. The SDK layers auth and retry on top of
// the provided transport so every outbound API call passes through it.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *Client) { c.transport = rt }
}

// WithTimeout sets the per-request HTTP timeout. Default is 30s.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) { c.timeout = timeout }
}

// WithRetryCount sets the number of additional retry attempts on 429 and 5xx
// responses. Default is 0 (no retries). The consuming service controls this;
// the SDK does not retry by default.
func WithRetryCount(n int) Option {
	return func(c *Client) { c.retryCount = n }
}

// withRetrySleep overrides the retry sleep function. Used in tests to avoid real delays.
func withRetrySleep(fn func(ctx context.Context, d time.Duration) bool) Option {
	return func(c *Client) { c.retrySleep = fn }
}

// NewClient creates a new Foursquare Places API client. At minimum, supply [WithAPIKey].
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// buildHTTPClient constructs the layered http.Client for a request.
// Transport stack (outermost to innermost):
//
//	authTransport (Authorization: Bearer + X-Places-Api-Version headers)
//	  -> retryTransport (if retryCount > 0)
//	     -> caller-supplied transport or http.DefaultTransport
func (c *Client) buildHTTPClient() *http.Client {
	var base = http.DefaultTransport
	if c.transport != nil {
		base = c.transport
	} else if c.httpClient != nil && c.httpClient.Transport != nil {
		base = c.httpClient.Transport
	}

	var transport = base
	if c.retryCount > 0 {
		transport = &retryTransport{next: transport, maxRetries: c.retryCount, sleep: c.retrySleep}
	}
	transport = &authTransport{apiKey: c.apiKey, next: transport}

	return &http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}
}

// do executes an HTTP request using the layered client.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", userAgent())
	hc := c.buildHTTPClient()
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	return resp, nil
}

// authTransport injects the Bearer token and required API version header.
type authTransport struct {
	apiKey string
	next   http.RoundTripper
}

func (tr *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+tr.apiKey)
	clone.Header.Set("X-Places-Api-Version", apiVersion)
	return tr.next.RoundTrip(clone)
}

func userAgent() string {
	const base = "foursquare-go"
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return base
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/way-platform/foursquare-go" && dep.Version != "" {
			return base + "/" + strings.TrimPrefix(dep.Version, "v")
		}
	}
	if info.Main.Path == "github.com/way-platform/foursquare-go" && info.Main.Version != "" {
		return base + "/" + strings.TrimPrefix(info.Main.Version, "v")
	}
	return base
}
