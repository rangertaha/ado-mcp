// SPDX-License-Identifier: MIT

package serviceendpoint

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the service endpoint tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "serviceendpoint_list", Title: "List service endpoints",
		Description: "List the service endpoints (service connections) in a project."}, svc.list)
	server.Register(s, server.ToolDef{Name: "serviceendpoint_get", Title: "Get service endpoint",
		Description: "Get a single service endpoint by ID within a project."}, svc.get)
	server.Register(s, server.ToolDef{Name: "serviceendpoint_list_types", Title: "List service endpoint types",
		Description: "List the available service endpoint types."}, svc.listTypes)
}

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// ListInput identifies the project whose service endpoints to list.
type ListInput struct {
	Project string `json:"project" jsonschema:"the project ID or name"`
}

// GetInput identifies a service endpoint to retrieve.
type GetInput struct {
	Project    string `json:"project" jsonschema:"the project ID or name"`
	EndpointID string `json:"endpointId" jsonschema:"the service endpoint ID"`
}

func (s *service) list(ctx context.Context, _ *mcp.CallToolRequest, in ListInput) (*mcp.CallToolResult, server.ListResult[ServiceEndpoint], error) {
	out, err := s.ListEndpoints(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) get(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, *ServiceEndpoint, error) {
	out, err := s.GetEndpoint(ctx, in.Project, in.EndpointID)
	return nil, out, err
}

func (s *service) listTypes(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[ServiceEndpointType], error) {
	out, err := s.ListTypes(ctx)
	return nil, server.List(out), err
}
