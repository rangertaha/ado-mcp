// SPDX-License-Identifier: MIT

package profile

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the profile tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "profile_get_me", Title: "Get my profile",
		Description: "Retrieve the profile of the authenticated user."}, svc.getMe)
}

// --- Tool input types ---

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// --- Tool handlers ---

func (s *service) getMe(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, *Profile, error) {
	out, err := s.GetMe(ctx)
	return nil, out, err
}
