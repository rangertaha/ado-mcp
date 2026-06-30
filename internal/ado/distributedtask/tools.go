// SPDX-License-Identifier: MIT

package distributedtask

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the distributed task tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{
		Name:        "dtask_list_variable_groups",
		Title:       "List variable groups",
		Description: "List the variable groups within a project.",
	}, svc.listVariableGroups)

	server.Register(s, server.ToolDef{
		Name:        "dtask_get_variable_group",
		Title:       "Get variable group",
		Description: "Get a single variable group by its ID.",
	}, svc.getVariableGroup)

	server.Register(s, server.ToolDef{
		Name:        "dtask_create_variable_group",
		Title:       "Create variable group",
		Description: "Create a variable group within a project from a raw body.",
		Write:       true,
	}, svc.createVariableGroup)

	server.Register(s, server.ToolDef{
		Name:        "dtask_list_environments",
		Title:       "List environments",
		Description: "List the deployment environments within a project.",
	}, svc.listEnvironments)

	server.Register(s, server.ToolDef{
		Name:        "dtask_get_environment",
		Title:       "Get environment",
		Description: "Get a single deployment environment by its ID.",
	}, svc.getEnvironment)

	server.Register(s, server.ToolDef{
		Name:        "dtask_create_environment",
		Title:       "Create environment",
		Description: "Create a deployment environment within a project.",
		Write:       true,
	}, svc.createEnvironment)

	server.Register(s, server.ToolDef{
		Name:        "dtask_list_agent_pools",
		Title:       "List agent pools",
		Description: "List the organization-level agent pools.",
	}, svc.listAgentPools)

	server.Register(s, server.ToolDef{
		Name:        "dtask_list_secure_files",
		Title:       "List secure files",
		Description: "List the secure files within a project.",
	}, svc.listSecureFiles)
}

// ListVariableGroupsInput selects the project whose variable groups are listed.
type ListVariableGroupsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

func (s *service) listVariableGroups(ctx context.Context, _ *mcp.CallToolRequest, in ListVariableGroupsInput) (*mcp.CallToolResult, server.ListResult[VariableGroup], error) {
	out, err := s.ListVariableGroups(ctx, in.Project)
	return nil, server.List(out), err
}

// GetVariableGroupInput identifies a single variable group.
type GetVariableGroupInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	GroupID int    `json:"groupId" jsonschema:"the variable group ID"`
}

func (s *service) getVariableGroup(ctx context.Context, _ *mcp.CallToolRequest, in GetVariableGroupInput) (*mcp.CallToolResult, *VariableGroup, error) {
	out, err := s.GetVariableGroup(ctx, in.Project, in.GroupID)
	return nil, out, err
}

// CreateVariableGroupInput carries the project and raw variable group body.
type CreateVariableGroupInput struct {
	Project string         `json:"project" jsonschema:"project name or ID"`
	Body    map[string]any `json:"body" jsonschema:"the variable group definition body"`
}

func (s *service) createVariableGroup(ctx context.Context, _ *mcp.CallToolRequest, in CreateVariableGroupInput) (*mcp.CallToolResult, *VariableGroup, error) {
	out, err := s.CreateVariableGroup(ctx, in.Project, in.Body)
	return nil, out, err
}

// ListEnvironmentsInput selects the project whose environments are listed.
type ListEnvironmentsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

func (s *service) listEnvironments(ctx context.Context, _ *mcp.CallToolRequest, in ListEnvironmentsInput) (*mcp.CallToolResult, server.ListResult[Environment], error) {
	out, err := s.ListEnvironments(ctx, in.Project)
	return nil, server.List(out), err
}

// GetEnvironmentInput identifies a single deployment environment.
type GetEnvironmentInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	EnvironmentID int    `json:"environmentId" jsonschema:"the environment ID"`
}

func (s *service) getEnvironment(ctx context.Context, _ *mcp.CallToolRequest, in GetEnvironmentInput) (*mcp.CallToolResult, *Environment, error) {
	out, err := s.GetEnvironment(ctx, in.Project, in.EnvironmentID)
	return nil, out, err
}

// CreateEnvironmentInput names the environment to create within a project.
type CreateEnvironmentInput struct {
	Project     string `json:"project" jsonschema:"project name or ID"`
	Name        string `json:"name" jsonschema:"the name of the environment"`
	Description string `json:"description,omitempty" jsonschema:"an optional description for the environment"`
}

func (s *service) createEnvironment(ctx context.Context, _ *mcp.CallToolRequest, in CreateEnvironmentInput) (*mcp.CallToolResult, *Environment, error) {
	out, err := s.CreateEnvironment(ctx, in.Project, in.Name, in.Description)
	return nil, out, err
}

// ListAgentPoolsInput takes no arguments.
type ListAgentPoolsInput struct{}

func (s *service) listAgentPools(ctx context.Context, _ *mcp.CallToolRequest, _ ListAgentPoolsInput) (*mcp.CallToolResult, server.ListResult[AgentPool], error) {
	out, err := s.ListAgentPools(ctx)
	return nil, server.List(out), err
}

// ListSecureFilesInput selects the project whose secure files are listed.
type ListSecureFilesInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

func (s *service) listSecureFiles(ctx context.Context, _ *mcp.CallToolRequest, in ListSecureFilesInput) (*mcp.CallToolResult, server.ListResult[SecureFile], error) {
	out, err := s.ListSecureFiles(ctx, in.Project)
	return nil, server.List(out), err
}
