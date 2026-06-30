// SPDX-License-Identifier: MIT

package graph

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Graph/Identity tools to the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "graph_list_users", Title: "List users",
		Description: "List the user subjects in the organization graph."}, svc.listUsers)
	server.Register(s, server.ToolDef{Name: "graph_list_groups", Title: "List groups",
		Description: "List the group subjects in the organization graph."}, svc.listGroups)
	server.Register(s, server.ToolDef{Name: "graph_list_memberships", Title: "List memberships",
		Description: "List the memberships of a subject (optionally up to containers or down to members)."}, svc.listMemberships)
	server.Register(s, server.ToolDef{Name: "graph_add_membership", Title: "Add membership",
		Description: "Create a membership relating a subject to a container.", Write: true, Idempotent: true}, svc.addMembership)
	server.Register(s, server.ToolDef{Name: "graph_remove_membership", Title: "Remove membership",
		Description: "Remove a membership relating a subject to a container.", Write: true, Destructive: true, Idempotent: true}, svc.removeMembership)
	server.Register(s, server.ToolDef{Name: "graph_get_descriptor", Title: "Get descriptor",
		Description: "Resolve a storage key (subject id) to its graph descriptor."}, svc.getDescriptor)
}

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

func (s *service) listUsers(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[GraphUser], error) {
	out, err := s.ListUsers(ctx)
	return nil, server.List(out), err
}

func (s *service) listGroups(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[GraphGroup], error) {
	out, err := s.ListGroups(ctx)
	return nil, server.List(out), err
}

// ListMembershipsInput selects the subject and optional traversal direction.
type ListMembershipsInput struct {
	SubjectDescriptor string `json:"subjectDescriptor" jsonschema:"the descriptor of the subject (user or group) whose memberships are listed"`
	Direction         string `json:"direction,omitempty" jsonschema:"traversal direction: 'up' for containers, 'down' for members (optional)"`
}

func (s *service) listMemberships(ctx context.Context, _ *mcp.CallToolRequest, in ListMembershipsInput) (*mcp.CallToolResult, server.ListResult[Membership], error) {
	out, err := s.ListMemberships(ctx, in.SubjectDescriptor, in.Direction)
	return nil, server.List(out), err
}

// MembershipInput identifies the subject and container of a membership.
type MembershipInput struct {
	SubjectDescriptor   string `json:"subjectDescriptor" jsonschema:"the descriptor of the member subject (user or group)"`
	ContainerDescriptor string `json:"containerDescriptor" jsonschema:"the descriptor of the container group"`
}

func (s *service) addMembership(ctx context.Context, _ *mcp.CallToolRequest, in MembershipInput) (*mcp.CallToolResult, *Membership, error) {
	out, err := s.AddMembership(ctx, in.SubjectDescriptor, in.ContainerDescriptor)
	return nil, out, err
}

func (s *service) removeMembership(ctx context.Context, _ *mcp.CallToolRequest, in MembershipInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.RemoveMembership(ctx, in.SubjectDescriptor, in.ContainerDescriptor); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

// GetDescriptorInput selects the storage key to resolve.
type GetDescriptorInput struct {
	StorageKey string `json:"storageKey" jsonschema:"the storage key (subject id) to resolve to a graph descriptor"`
}

func (s *service) getDescriptor(ctx context.Context, _ *mcp.CallToolRequest, in GetDescriptorInput) (*mcp.CallToolResult, *Descriptor, error) {
	out, err := s.GetDescriptor(ctx, in.StorageKey)
	return nil, out, err
}
