// SPDX-License-Identifier: MIT

package memberentitlement

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the member entitlement tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "entitlement_list_users", Title: "List user entitlements",
		Description: "List user access-level entitlements (licensing) for the organization."}, svc.listUsers)
	server.Register(s, server.ToolDef{Name: "entitlement_get_user", Title: "Get user entitlement",
		Description: "Get the access-level entitlement for a single user by entitlement id."}, svc.getUser)
	server.Register(s, server.ToolDef{Name: "entitlement_list_groups", Title: "List group entitlements",
		Description: "List group entitlements (licensing rules) for the organization."}, svc.listGroups)
}

// --- Tool input types ---

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// GetUserInput identifies the user entitlement to retrieve.
type GetUserInput struct {
	UserID string `json:"userId" jsonschema:"the user entitlement id"`
}

// --- Tool handlers ---

func (s *service) listUsers(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, *UserEntitlementList, error) {
	out, err := s.ListUsers(ctx)
	return nil, out, err
}

func (s *service) getUser(ctx context.Context, _ *mcp.CallToolRequest, in GetUserInput) (*mcp.CallToolResult, *UserEntitlement, error) {
	out, err := s.GetUser(ctx, in.UserID)
	return nil, out, err
}

func (s *service) listGroups(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[GroupEntitlement], error) {
	out, err := s.ListGroups(ctx)
	return nil, server.List(out), err
}
