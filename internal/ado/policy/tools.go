// SPDX-License-Identifier: MIT

package policy

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the policy tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{
		Name:        "policy_list_configurations",
		Title:       "List policy configurations",
		Description: "List the policy configurations for a project.",
	}, svc.listConfigurations)

	server.Register(s, server.ToolDef{
		Name:        "policy_get_configuration",
		Title:       "Get policy configuration",
		Description: "Get a single policy configuration by its ID.",
	}, svc.getConfiguration)

	server.Register(s, server.ToolDef{
		Name:        "policy_list_types",
		Title:       "List policy types",
		Description: "List the policy types available within a project.",
	}, svc.listTypes)

	server.Register(s, server.ToolDef{
		Name:        "policy_create_configuration",
		Title:       "Create policy configuration",
		Description: "Create a new policy configuration within a project.",
		Write:       true,
	}, svc.createConfiguration)
}

// ListConfigurationsInput selects the project whose policy configurations are listed.
type ListConfigurationsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

func (s *service) listConfigurations(ctx context.Context, _ *mcp.CallToolRequest, in ListConfigurationsInput) (*mcp.CallToolResult, server.ListResult[PolicyConfiguration], error) {
	out, err := s.ListConfigurations(ctx, in.Project)
	return nil, server.List(out), err
}

// GetConfigurationInput identifies a single policy configuration.
type GetConfigurationInput struct {
	Project         string `json:"project" jsonschema:"project name or ID"`
	ConfigurationID int    `json:"configurationId" jsonschema:"the policy configuration ID"`
}

func (s *service) getConfiguration(ctx context.Context, _ *mcp.CallToolRequest, in GetConfigurationInput) (*mcp.CallToolResult, *PolicyConfiguration, error) {
	out, err := s.GetConfiguration(ctx, in.Project, in.ConfigurationID)
	return nil, out, err
}

// ListTypesInput selects the project whose policy types are listed.
type ListTypesInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

func (s *service) listTypes(ctx context.Context, _ *mcp.CallToolRequest, in ListTypesInput) (*mcp.CallToolResult, server.ListResult[PolicyType], error) {
	out, err := s.ListTypes(ctx, in.Project)
	return nil, server.List(out), err
}

// CreateConfigurationInput carries the project and raw policy configuration body.
type CreateConfigurationInput struct {
	Project string         `json:"project" jsonschema:"project name or ID"`
	Body    map[string]any `json:"body" jsonschema:"the policy configuration body (type, settings, isEnabled, isBlocking)"`
}

func (s *service) createConfiguration(ctx context.Context, _ *mcp.CallToolRequest, in CreateConfigurationInput) (*mcp.CallToolResult, *PolicyConfiguration, error) {
	out, err := s.CreateConfiguration(ctx, in.Project, in.Body)
	return nil, out, err
}
