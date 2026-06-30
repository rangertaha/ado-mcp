// SPDX-License-Identifier: MIT

package search

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the search tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "search_code", Title: "Search code",
		Description: "Full-text search across source code in the organization, optionally scoped to a project."}, svc.searchCode)
	server.Register(s, server.ToolDef{Name: "search_work_items", Title: "Search work items",
		Description: "Full-text search across work items in the organization, optionally scoped to a project."}, svc.searchWorkItems)
	server.Register(s, server.ToolDef{Name: "search_wiki", Title: "Search wiki",
		Description: "Full-text search across wiki pages in the organization, optionally scoped to a project."}, svc.searchWiki)
}

// --- Tool input types ---

// SearchInput is the common input for all search tools.
type SearchInput struct {
	SearchText string `json:"searchText" jsonschema:"the text to search for"`
	Project    string `json:"project,omitempty" jsonschema:"name of the project to scope the search to (optional)"`
	Top        int    `json:"top,omitempty" jsonschema:"maximum number of results to return (optional, default 50)"`
}

// --- Tool handlers ---

func (s *service) searchCode(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, *CodeSearchResponse, error) {
	out, err := s.SearchCode(ctx, in.SearchText, in.Project, in.Top)
	return nil, out, err
}

func (s *service) searchWorkItems(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, *WorkItemSearchResponse, error) {
	out, err := s.SearchWorkItems(ctx, in.SearchText, in.Project, in.Top)
	return nil, out, err
}

func (s *service) searchWiki(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, *WikiSearchResponse, error) {
	out, err := s.SearchWiki(ctx, in.SearchText, in.Project, in.Top)
	return nil, out, err
}
