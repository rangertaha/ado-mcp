// SPDX-License-Identifier: MIT

package security

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Security toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "security_list_namespaces", Title: "List security namespaces",
		Description: "List the security namespaces in the organization, optionally filtered by namespace ID."}, svc.listNamespaces)
	server.Register(s, server.ToolDef{Name: "security_list_access_control_lists", Title: "List access control lists",
		Description: "List the access control lists for a security namespace, optionally filtered by token."}, svc.listAccessControlLists)
	server.Register(s, server.ToolDef{Name: "security_set_access_control_entries", Title: "Set access control entries",
		Description: "Add or update access control entries on the ACL for a token in a security namespace.",
		Write:       true}, svc.setAccessControlEntries)
	server.Register(s, server.ToolDef{Name: "security_remove_access_control_lists", Title: "Remove access control lists",
		Description: "Remove the access control lists for the given tokens in a security namespace.",
		Write:       true, Destructive: true}, svc.removeAccessControlLists)
}

// --- Tool input types ---

// ListNamespacesInput optionally filters security namespaces by ID.
type ListNamespacesInput struct {
	NamespaceID string `json:"namespaceId,omitempty" jsonschema:"security namespace ID to filter by (optional)"`
}

// ListAccessControlListsInput selects the access control lists to return.
type ListAccessControlListsInput struct {
	NamespaceID         string `json:"namespaceId" jsonschema:"security namespace ID"`
	Token               string `json:"token,omitempty" jsonschema:"security token to filter the ACLs (optional)"`
	IncludeExtendedInfo bool   `json:"includeExtendedInfo,omitempty" jsonschema:"include extended permission info (optional)"`
}

// SetAccessControlEntriesInput passes through the access control entries to set.
type SetAccessControlEntriesInput struct {
	NamespaceID string         `json:"namespaceId" jsonschema:"security namespace ID"`
	Body        map[string]any `json:"body" jsonschema:"request body with {token: string, merge: bool, accessControlEntries: array of ACEs}"`
}

// RemoveAccessControlListsInput selects the ACLs to remove by token.
type RemoveAccessControlListsInput struct {
	NamespaceID string `json:"namespaceId" jsonschema:"security namespace ID"`
	Tokens      string `json:"tokens" jsonschema:"one or more security tokens whose ACLs should be removed"`
}

// --- Handlers ---

func (s *service) listNamespaces(ctx context.Context, _ *mcp.CallToolRequest, in ListNamespacesInput) (*mcp.CallToolResult, server.ListResult[SecurityNamespace], error) {
	out, err := s.ListNamespaces(ctx, in.NamespaceID)
	return nil, server.List(out), err
}

func (s *service) listAccessControlLists(ctx context.Context, _ *mcp.CallToolRequest, in ListAccessControlListsInput) (*mcp.CallToolResult, server.ListResult[AccessControlList], error) {
	out, err := s.ListAccessControlLists(ctx, in.NamespaceID, in.Token, in.IncludeExtendedInfo)
	return nil, server.List(out), err
}

func (s *service) setAccessControlEntries(ctx context.Context, _ *mcp.CallToolRequest, in SetAccessControlEntriesInput) (*mcp.CallToolResult, server.ListResult[any], error) {
	out, err := s.SetAccessControlEntries(ctx, in.NamespaceID, in.Body)
	return nil, server.List(out), err
}

func (s *service) removeAccessControlLists(ctx context.Context, _ *mcp.CallToolRequest, in RemoveAccessControlListsInput) (*mcp.CallToolResult, *struct{}, error) {
	err := s.RemoveAccessControlLists(ctx, in.NamespaceID, in.Tokens)
	if err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}
