// SPDX-License-Identifier: MIT

// Package client provides a small, dependency-free HTTP client for talking to
// JSON REST APIs. It is the single transport layer shared by every Azure DevOps
// service wrapper and by the 7pace Timetracker integration.
//
// The client is intentionally generic: callers describe a request (method,
// path, query, body) and supply a destination for the decoded JSON response.
// Authentication is pluggable via the Authorizer interface so the same code
// serves both Azure DevOps (PAT Basic auth) and 7pace (bearer token).
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// defaultTimeout bounds a single HTTP request.
const defaultTimeout = 60 * time.Second

// Authorizer applies authentication credentials to an outgoing request.
type Authorizer interface {
	// Authorize mutates the request to carry authentication.
	Authorize(*http.Request)
}

// Client is a reusable JSON REST client bound to a base URL and an Authorizer.
// A Client is safe for concurrent use by multiple goroutines.
type Client struct {
	base       *url.URL
	http       *http.Client
	auth       Authorizer
	userAgent  string
	apiVersion string // default api-version query param; "" means none
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom *http.Client (e.g. for testing or proxies).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

// WithUserAgent sets the User-Agent header sent on every request.
func WithUserAgent(ua string) Option { return func(c *Client) { c.userAgent = ua } }

// WithAPIVersion sets a default "api-version" query parameter applied to every
// request that does not specify its own. Azure DevOps requires this; 7pace
// encodes the version in the path and should leave it empty.
func WithAPIVersion(v string) Option { return func(c *Client) { c.apiVersion = v } }

// New creates a Client for the given base URL and Authorizer.
func New(baseURL string, auth Authorizer, opts ...Option) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL %q: %w", baseURL, err)
	}
	c := &Client{
		base:      u,
		http:      &http.Client{Timeout: defaultTimeout},
		auth:      auth,
		userAgent: "ado-mcp",
	}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Request describes a single REST call.
type Request struct {
	// Method is the HTTP method (GET, POST, PATCH, PUT, DELETE).
	Method string

	// Path is appended to the client's base URL. A leading slash is optional.
	Path string

	// Query holds URL query parameters. The "api-version" parameter, if not
	// present here, is added from the client default when set.
	Query url.Values

	// Body, when non-nil, is JSON-encoded and sent as the request body.
	// If Body already implements io.Reader it is sent verbatim.
	Body any

	// ContentType overrides the request Content-Type. Defaults to
	// "application/json" when a body is present.
	ContentType string

	// APIVersion overrides the client's default api-version for this request.
	APIVersion string

	// Header holds extra request headers (e.g. "If-Match" for conditional
	// updates). These are applied after defaults, so they take precedence.
	Header http.Header

	// Out, when non-nil, receives the JSON-decoded response body.
	Out any
}

// Response carries metadata about a completed request. The decoded body, if
// requested, is written to Request.Out.
type Response struct {
	StatusCode int
	Header     http.Header

	// ContinuationToken is the value of the x-ms-continuationtoken header used
	// by Azure DevOps to page through large result sets. Empty when absent.
	ContinuationToken string
}

// Do executes a request, decoding a successful JSON response into req.Out (when
// set) and returning a typed *APIError for non-2xx responses.
//
// If Azure DevOps rejects the requested api-version (HTTP 400, "version not
// supported"), Do negotiates a supported version from the server's response and
// retries the request once. This transparently handles endpoints that require a
// "-preview" api-version without each caller having to know the exact suffix.
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	// Marshal a JSON body once up front so the negotiation retry below can resend
	// it (from a fresh reader) without re-marshaling. Streaming (io.Reader) bodies
	// are left as-is and are not retried (see versionRetryable).
	if req.Body != nil {
		if _, isReader := req.Body.(io.Reader); !isReader {
			data, err := json.Marshal(req.Body)
			if err != nil {
				return nil, fmt.Errorf("encoding request body: %w", err)
			}
			ct := req.ContentType
			if ct == "" {
				ct = "application/json"
			}
			req.Body = preEncoded{data: data, contentType: ct}
			req.ContentType = ct
		}
	}

	out, body, httpReq, err := c.do(ctx, req)
	if err != nil {
		return out, err
	}

	if out.StatusCode < 200 || out.StatusCode >= 300 {
		// One-shot api-version negotiation on a version-mismatch 400.
		if alt := negotiateAPIVersion(c.effectiveVersion(req), out.Header, body); alt != "" && versionRetryable(req) {
			req.APIVersion = alt
			out, body, httpReq, err = c.do(ctx, req)
			if err != nil {
				return out, err
			}
		}
	}

	if out.StatusCode < 200 || out.StatusCode >= 300 {
		return out, parseAPIError(req.Method, httpReq.URL, out.StatusCode, body)
	}

	// Raw (non-JSON) responses, e.g. build logs, are captured verbatim.
	if rb, ok := req.Out.(*RawBody); ok {
		rb.Bytes = body
		rb.ContentType = out.Header.Get("Content-Type")
		return out, nil
	}

	if req.Out != nil && len(body) > 0 {
		if err := json.Unmarshal(body, req.Out); err != nil {
			return out, fmt.Errorf("decoding %s %s response: %w", req.Method, httpReq.URL.Path, err)
		}
	}
	return out, nil
}

// do performs a single HTTP round-trip and returns the response metadata, the
// raw body, and the *http.Request that was sent (for error context).
func (c *Client) do(ctx context.Context, req Request) (*Response, []byte, *http.Request, error) {
	httpReq, err := c.buildRequest(ctx, req)
	if err != nil {
		return nil, nil, nil, err
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, nil, httpReq, fmt.Errorf("%s %s: %w", req.Method, httpReq.URL.Path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, httpReq, fmt.Errorf("reading response body: %w", err)
	}
	out := &Response{
		StatusCode:        resp.StatusCode,
		Header:            resp.Header,
		ContinuationToken: resp.Header.Get("x-ms-continuationtoken"),
	}
	return out, body, httpReq, nil
}

// effectiveVersion returns the api-version that a request will actually send.
// This must mirror buildRequest's precedence (req.APIVersion, then an explicit
// api-version in req.Query, then the client default) so version negotiation
// compares against the value really transmitted.
func (c *Client) effectiveVersion(req Request) string {
	if req.APIVersion != "" {
		return req.APIVersion
	}
	if v := req.Query.Get("api-version"); v != "" {
		return v
	}
	return c.apiVersion
}

// versionRetryable reports whether a request can be safely re-sent for version
// negotiation. Requests with a streaming (io.Reader) body cannot, because the
// reader would already be consumed.
func versionRetryable(req Request) bool {
	if req.Body == nil {
		return true
	}
	_, isReader := req.Body.(io.Reader)
	return !isReader
}

// RawBody captures an undecoded response body. Pass a *RawBody as Request.Out
// to receive the raw bytes (and content type) instead of JSON decoding, for
// endpoints that return plain text such as build/pipeline logs.
type RawBody struct {
	Bytes       []byte
	ContentType string
}

// String returns the raw body as a string.
func (r *RawBody) String() string { return string(r.Bytes) }

// buildRequest assembles an *http.Request from a Request.
func (c *Client) buildRequest(ctx context.Context, req Request) (*http.Request, error) {
	u := *c.base
	// req.Path arrives already percent-encoded (callers url.PathEscape their
	// segments). Assign it as RawPath and store the decoded form in Path so
	// url.URL.String() emits it verbatim instead of re-encoding the '%' (which
	// would double-encode spaces etc. to %2520).
	joined := joinPath(c.base.Path, req.Path)
	if dec, err := url.PathUnescape(joined); err == nil {
		u.Path = dec
		u.RawPath = joined
	} else {
		u.Path = joined
	}

	// Clone the caller's query map before mutating it: Request is passed by value
	// but the url.Values map header is shared, so writing api-version directly
	// would leak into a map the caller still owns (and the negotiation retry
	// would overwrite it).
	q := url.Values{}
	for k, v := range req.Query {
		q[k] = v
	}
	if v := req.APIVersion; v != "" {
		q.Set("api-version", v)
	} else if c.apiVersion != "" && q.Get("api-version") == "" {
		q.Set("api-version", c.apiVersion)
	}
	u.RawQuery = q.Encode()

	body, contentType, err := encodeBody(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	if c.userAgent != "" {
		httpReq.Header.Set("User-Agent", c.userAgent)
	}
	if c.auth != nil {
		c.auth.Authorize(httpReq)
	}
	// Caller-supplied headers take precedence over defaults.
	for k, vs := range req.Header {
		httpReq.Header.Del(k)
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}
	return httpReq, nil
}

// preEncoded is a request body already marshaled to bytes. Do replaces a JSON
// body with this so the api-version negotiation retry can resend it from a
// fresh reader without marshaling a second time.
type preEncoded struct {
	data        []byte
	contentType string
}

// encodeBody turns Request.Body into an io.Reader and resolves the Content-Type.
func encodeBody(req Request) (io.Reader, string, error) {
	if req.Body == nil {
		return nil, "", nil
	}
	if pe, ok := req.Body.(preEncoded); ok {
		return bytes.NewReader(pe.data), pe.contentType, nil
	}
	if r, ok := req.Body.(io.Reader); ok {
		ct := req.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		return r, ct, nil
	}
	data, err := json.Marshal(req.Body)
	if err != nil {
		return nil, "", fmt.Errorf("encoding request body: %w", err)
	}
	ct := req.ContentType
	if ct == "" {
		ct = "application/json"
	}
	return bytes.NewReader(data), ct, nil
}

// joinPath joins a base path and a relative path with exactly one separator.
func joinPath(base, rel string) string {
	base = strings.TrimRight(base, "/")
	rel = strings.TrimLeft(rel, "/")
	switch {
	case base == "" && rel == "":
		return "/"
	case rel == "":
		return base
	case base == "":
		return "/" + rel
	default:
		return base + "/" + rel
	}
}
