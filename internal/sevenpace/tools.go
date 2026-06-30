// SPDX-License-Identifier: MIT

package sevenpace

import (
	"context"
	"net/url"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the 7pace Timetracker toolset. The 7pace Reporting OData API is
// read-only (and requires a Service Account configured in 7pace settings), so
// all tools here are reads.
func Register(s *server.Server, c *Client) {
	s.NoteToolset(Name)

	server.Register(s, server.ToolDef{
		Name:        "sevenpace_list_worklogs",
		Title:       "List worklogs",
		Description: "List 7pace worklogs (the workLogsOnly entity). Filter with an OData $filter such as \"Timestamp ge 2026-01-01T00:00:00Z\".",
	}, c.listWorkLogs)

	server.Register(s, server.ToolDef{
		Name:        "sevenpace_list_worklog_workitems",
		Title:       "List worklogs with work items",
		Description: "List 7pace worklogs joined with their Azure DevOps work items (the workLogsWorkItems entity). Supports an OData $filter.",
	}, c.listWorkLogWorkItems)

	server.Register(s, server.ToolDef{
		Name:        "sevenpace_list_work_items",
		Title:       "List work items with tracked time",
		Description: "List work items with their rolled-up tracked time (the workItems entity). Supports an OData $filter.",
	}, c.listWorkItems)

	server.Register(s, server.ToolDef{
		Name:        "sevenpace_list_budgets",
		Title:       "List budgets",
		Description: "List the configured 7pace budgets.",
	}, c.listBudgets)

	server.Register(s, server.ToolDef{
		Name:        "sevenpace_query",
		Title:       "Query 7pace OData",
		Description: "Run a raw GET against the 7pace Reporting OData API and return the response body. Use for entity sets or options not covered by the other tools, or path \"$metadata\" to inspect the schema. Entity sets include workLogsOnly, workLogsWorkItems, workItems, budgets.",
	}, c.query)
}

// --- Inputs / outputs ---

// FilterInput filters and pages an OData entity set.
type FilterInput struct {
	Filter string `json:"filter,omitempty" jsonschema:"optional OData $filter expression"`
	Top    int    `json:"top,omitempty" jsonschema:"maximum number of records to return (optional)"`
}

// TopInput pages an OData entity set.
type TopInput struct {
	Top int `json:"top,omitempty" jsonschema:"maximum number of records to return (optional)"`
}

// QueryInput is a raw OData request.
type QueryInput struct {
	Path    string `json:"path" jsonschema:"OData path, e.g. workLogsOnly, workItems, or $metadata"`
	Filter  string `json:"filter,omitempty" jsonschema:"optional OData $filter expression"`
	Select  string `json:"select,omitempty" jsonschema:"optional OData $select (comma-separated fields)"`
	OrderBy string `json:"orderby,omitempty" jsonschema:"optional OData $orderby expression"`
	Top     int    `json:"top,omitempty" jsonschema:"optional OData $top"`
}

// QueryOutput wraps a raw OData response body.
type QueryOutput struct {
	Body string `json:"body" jsonschema:"the raw response body"`
}

// --- Handlers ---

func (c *Client) listWorkLogs(ctx context.Context, _ *mcp.CallToolRequest, in FilterInput) (*mcp.CallToolResult, server.ListResult[Record], error) {
	out, err := c.ListWorkLogs(ctx, in.Filter, in.Top)
	return nil, server.List(out), err
}

func (c *Client) listWorkLogWorkItems(ctx context.Context, _ *mcp.CallToolRequest, in FilterInput) (*mcp.CallToolResult, server.ListResult[Record], error) {
	out, err := c.ListWorkLogsWithWorkItems(ctx, in.Filter, in.Top)
	return nil, server.List(out), err
}

func (c *Client) listWorkItems(ctx context.Context, _ *mcp.CallToolRequest, in FilterInput) (*mcp.CallToolResult, server.ListResult[Record], error) {
	out, err := c.ListWorkItems(ctx, in.Filter, in.Top)
	return nil, server.List(out), err
}

func (c *Client) listBudgets(ctx context.Context, _ *mcp.CallToolRequest, in TopInput) (*mcp.CallToolResult, server.ListResult[Record], error) {
	out, err := c.ListBudgets(ctx, in.Top)
	return nil, server.List(out), err
}

func (c *Client) query(ctx context.Context, _ *mcp.CallToolRequest, in QueryInput) (*mcp.CallToolResult, *QueryOutput, error) {
	opts := url.Values{}
	if in.Filter != "" {
		opts.Set("$filter", in.Filter)
	}
	if in.Select != "" {
		opts.Set("$select", in.Select)
	}
	if in.OrderBy != "" {
		opts.Set("$orderby", in.OrderBy)
	}
	if in.Top > 0 {
		opts.Set("$top", strconv.Itoa(in.Top))
	}
	body, err := c.QueryRaw(ctx, in.Path, opts)
	if err != nil {
		return nil, nil, err
	}
	return nil, &QueryOutput{Body: body}, nil
}
