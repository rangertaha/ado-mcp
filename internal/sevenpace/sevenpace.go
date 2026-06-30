// SPDX-License-Identifier: MIT

// Package sevenpace integrates 7pace Timetracker (an Appfire product) for Azure
// DevOps via its Reporting OData API (e.g.
// https://{org}.timehub.7pace.com/api/odata/v3.2), authenticated with a bearer
// token. It exposes worklog and work-item reporting queries as MCP tools.
//
// The OData Reporting API is read-only and requires a Service Account to be
// configured in 7pace Timetracker settings; without it the API returns HTTP 403
// "ServiceAccountNotSet". The base URL comes from configuration
// (config.SevenPaceBaseURL), which honors the SEVENPACE_API_BASE override.
package sevenpace

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "sevenpace"

// OData entity sets exposed by the 7pace Reporting API.
const (
	entityWorkLogsOnly      = "workLogsOnly"      // worklogs without work-item details
	entityWorkLogsWorkItems = "workLogsWorkItems" // worklogs joined with their work items
	entityWorkItems         = "workItems"         // work items with rolled-up tracked time
	entityBudgets           = "budgets"           // budgets
)

// Client wraps the REST client for the 7pace OData API.
type Client struct {
	c *client.Client
}

// NewClient builds a 7pace client for the given OData API base URL and bearer
// token.
func NewClient(baseURL, token string, opts ...client.Option) (*Client, error) {
	auth := client.NewBearerAuthorizer(token)
	base := append([]client.Option{client.WithUserAgent("ado-mcp")}, opts...)
	c, err := client.New(baseURL, auth, base...)
	if err != nil {
		return nil, fmt.Errorf("creating 7pace client: %w", err)
	}
	return &Client{c: c}, nil
}

// Record is a single OData entity returned by the Reporting API. Field names
// vary by entity (e.g. Timestamp, Length, User, ActivityType for worklogs), so
// records are returned untyped rather than forcing a fixed schema.
type Record = map[string]any

// odataList is the standard OData collection envelope: {"value": [...]}.
type odataList struct {
	Value []Record `json:"value"`
}

// list queries an OData entity set with optional $filter and $top.
func (c *Client) list(ctx context.Context, entity, filter string, top int) ([]Record, error) {
	q := url.Values{}
	if filter != "" {
		q.Set("$filter", filter)
	}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var env odataList
	if err := c.c.GetJSON(ctx, "/"+entity, q, &env); err != nil {
		return nil, err
	}
	return env.Value, nil
}

// ListWorkLogs returns worklog records (the workLogsOnly entity set), optionally
// filtered by an OData $filter (e.g. "Timestamp ge 2026-01-01T00:00:00Z").
func (c *Client) ListWorkLogs(ctx context.Context, filter string, top int) ([]Record, error) {
	return c.list(ctx, entityWorkLogsOnly, filter, top)
}

// ListWorkLogsWithWorkItems returns worklogs joined with their work items
// (the workLogsWorkItems entity set).
func (c *Client) ListWorkLogsWithWorkItems(ctx context.Context, filter string, top int) ([]Record, error) {
	return c.list(ctx, entityWorkLogsWorkItems, filter, top)
}

// ListWorkItems returns work items with rolled-up tracked time.
func (c *Client) ListWorkItems(ctx context.Context, filter string, top int) ([]Record, error) {
	return c.list(ctx, entityWorkItems, filter, top)
}

// ListBudgets returns the configured budgets.
func (c *Client) ListBudgets(ctx context.Context, top int) ([]Record, error) {
	return c.list(ctx, entityBudgets, "", top)
}

// QueryRaw performs a GET against an arbitrary OData path (e.g. an entity set or
// "$metadata") with raw query options, returning the response body verbatim.
// It is the escape hatch for entity sets or options not covered by the typed
// methods and for schema discovery.
func (c *Client) QueryRaw(ctx context.Context, path string, opts url.Values) (string, error) {
	var raw client.RawBody
	if _, err := c.c.Do(ctx, client.Request{Method: "GET", Path: "/" + path, Query: opts, Out: &raw}); err != nil {
		return "", err
	}
	return raw.String(), nil
}
