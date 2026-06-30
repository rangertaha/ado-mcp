// SPDX-License-Identifier: MIT

package processes

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the processes tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{
		Name:        "processes_list",
		Title:       "List processes",
		Description: "List the work item tracking processes in the organization.",
	}, svc.list)

	server.Register(s, server.ToolDef{
		Name:        "processes_get",
		Title:       "Get process",
		Description: "Get a single process by its process type ID.",
	}, svc.get)

	server.Register(s, server.ToolDef{
		Name:        "processes_list_work_item_types",
		Title:       "List work item types",
		Description: "List the work item types defined within a process.",
	}, svc.listWorkItemTypes)

	server.Register(s, server.ToolDef{
		Name:        "processes_list_fields",
		Title:       "List fields",
		Description: "List the fields of a work item type within a process.",
	}, svc.listFields)

	server.Register(s, server.ToolDef{
		Name:        "processes_list_states",
		Title:       "List states",
		Description: "List the workflow states of a work item type within a process.",
	}, svc.listStates)

	server.Register(s, server.ToolDef{
		Name:        "processes_list_behaviors",
		Title:       "List behaviors",
		Description: "List the behaviors defined within a process.",
	}, svc.listBehaviors)
}

// ListInput takes no arguments.
type ListInput struct{}

func (s *service) list(ctx context.Context, _ *mcp.CallToolRequest, _ ListInput) (*mcp.CallToolResult, server.ListResult[Process], error) {
	out, err := s.List(ctx)
	return nil, server.List(out), err
}

// GetInput identifies a single process.
type GetInput struct {
	ProcessTypeID string `json:"processTypeId" jsonschema:"the process type ID"`
}

func (s *service) get(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, *Process, error) {
	out, err := s.Get(ctx, in.ProcessTypeID)
	return nil, out, err
}

// ListWorkItemTypesInput identifies the process whose work item types are listed.
type ListWorkItemTypesInput struct {
	ProcessTypeID string `json:"processTypeId" jsonschema:"the process type ID"`
}

func (s *service) listWorkItemTypes(ctx context.Context, _ *mcp.CallToolRequest, in ListWorkItemTypesInput) (*mcp.CallToolResult, server.ListResult[ProcessWorkItemType], error) {
	out, err := s.ListWorkItemTypes(ctx, in.ProcessTypeID)
	return nil, server.List(out), err
}

// ListFieldsInput identifies the work item type whose fields are listed.
type ListFieldsInput struct {
	ProcessTypeID string `json:"processTypeId" jsonschema:"the process type ID"`
	WitRefName    string `json:"witRefName" jsonschema:"the work item type reference name"`
}

func (s *service) listFields(ctx context.Context, _ *mcp.CallToolRequest, in ListFieldsInput) (*mcp.CallToolResult, server.ListResult[ProcessField], error) {
	out, err := s.ListFields(ctx, in.ProcessTypeID, in.WitRefName)
	return nil, server.List(out), err
}

// ListStatesInput identifies the work item type whose states are listed.
type ListStatesInput struct {
	ProcessTypeID string `json:"processTypeId" jsonschema:"the process type ID"`
	WitRefName    string `json:"witRefName" jsonschema:"the work item type reference name"`
}

func (s *service) listStates(ctx context.Context, _ *mcp.CallToolRequest, in ListStatesInput) (*mcp.CallToolResult, server.ListResult[ProcessState], error) {
	out, err := s.ListStates(ctx, in.ProcessTypeID, in.WitRefName)
	return nil, server.List(out), err
}

// ListBehaviorsInput identifies the process whose behaviors are listed.
type ListBehaviorsInput struct {
	ProcessTypeID string `json:"processTypeId" jsonschema:"the process type ID"`
}

func (s *service) listBehaviors(ctx context.Context, _ *mcp.CallToolRequest, in ListBehaviorsInput) (*mcp.CallToolResult, server.ListResult[Behavior], error) {
	out, err := s.ListBehaviors(ctx, in.ProcessTypeID)
	return nil, server.List(out), err
}
