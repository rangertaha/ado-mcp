// SPDX-License-Identifier: MIT

package identities

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the legacy Identity tools to the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "identities_read", Title: "Read identities",
		Description: "Look up legacy identities by search filter (e.g. 'General', 'DisplayName', 'AccountName'), filter value or descriptors."}, svc.read)
}

// ReadInput selects how identities are looked up.
type ReadInput struct {
	SearchFilter string `json:"searchFilter,omitempty" jsonschema:"the search filter to apply, e.g. 'General', 'DisplayName' or 'AccountName' (optional)"`
	FilterValue  string `json:"filterValue,omitempty" jsonschema:"the value to match against the search filter (optional)"`
	Descriptors  string `json:"descriptors,omitempty" jsonschema:"a comma-separated list of identity descriptors to resolve (optional)"`
}

func (s *service) read(ctx context.Context, _ *mcp.CallToolRequest, in ReadInput) (*mcp.CallToolResult, server.ListResult[Identity], error) {
	out, err := s.Read(ctx, in.SearchFilter, in.FilterValue, in.Descriptors)
	return nil, server.List(out), err
}
