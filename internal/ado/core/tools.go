// SPDX-License-Identifier: MIT

package core

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Core toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{
		Name:        "core_list_projects",
		Title:       "List projects",
		Description: "List the team projects in the Azure DevOps organization.",
	}, svc.listProjects)

	server.Register(s, server.ToolDef{
		Name:        "core_get_project",
		Title:       "Get project",
		Description: "Get a single Azure DevOps project by name or ID.",
	}, svc.getProject)

	server.Register(s, server.ToolDef{
		Name:        "core_list_teams",
		Title:       "List teams",
		Description: "List the teams within a project.",
	}, svc.listTeams)

	server.Register(s, server.ToolDef{
		Name:        "core_get_team",
		Title:       "Get team",
		Description: "Get a single team within a project by name or ID.",
	}, svc.getTeam)

	server.Register(s, server.ToolDef{
		Name:        "core_list_team_members",
		Title:       "List team members",
		Description: "List the members of a team within a project.",
	}, svc.listTeamMembers)

	server.Register(s, server.ToolDef{
		Name:        "core_list_processes",
		Title:       "List processes",
		Description: "List the process templates (e.g. Agile, Scrum, CMMI) available to the organization.",
	}, svc.listProcesses)
}

// --- Tool input types (schemas are inferred from these structs) ---

// ListProjectsInput controls paging over projects.
type ListProjectsInput struct {
	Top  int `json:"top,omitempty" jsonschema:"maximum number of projects to return (optional)"`
	Skip int `json:"skip,omitempty" jsonschema:"number of projects to skip for paging (optional)"`
}

// ProjectInput identifies a single project.
type ProjectInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

// TeamMembersInput identifies a team within a project.
type TeamMembersInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
}

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// --- Tool handlers ---

func (s *service) listProjects(ctx context.Context, _ *mcp.CallToolRequest, in ListProjectsInput) (*mcp.CallToolResult, server.ListResult[Project], error) {
	out, err := s.ListProjects(ctx, in.Top, in.Skip)
	return nil, server.List(out), err
}

func (s *service) getProject(ctx context.Context, _ *mcp.CallToolRequest, in ProjectInput) (*mcp.CallToolResult, *Project, error) {
	out, err := s.GetProject(ctx, in.Project)
	return nil, out, err
}

func (s *service) getTeam(ctx context.Context, _ *mcp.CallToolRequest, in TeamMembersInput) (*mcp.CallToolResult, *Team, error) {
	out, err := s.GetTeam(ctx, in.Project, in.Team)
	return nil, out, err
}

func (s *service) listTeams(ctx context.Context, _ *mcp.CallToolRequest, in ProjectInput) (*mcp.CallToolResult, server.ListResult[Team], error) {
	out, err := s.ListTeams(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) listTeamMembers(ctx context.Context, _ *mcp.CallToolRequest, in TeamMembersInput) (*mcp.CallToolResult, server.ListResult[Member], error) {
	out, err := s.ListTeamMembers(ctx, in.Project, in.Team)
	return nil, server.List(out), err
}

func (s *service) listProcesses(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[Process], error) {
	out, err := s.ListProcesses(ctx)
	return nil, server.List(out), err
}
