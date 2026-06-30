// SPDX-License-Identifier: MIT

package dashboards

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the dashboards tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{
		Name:        "dashboards_list_dashboards",
		Title:       "List dashboards",
		Description: "List the dashboards for a team within a project.",
	}, svc.listDashboards)

	server.Register(s, server.ToolDef{
		Name:        "dashboards_get_dashboard",
		Title:       "Get dashboard",
		Description: "Get a single dashboard by its ID.",
	}, svc.getDashboard)

	server.Register(s, server.ToolDef{
		Name:        "dashboards_list_widgets",
		Title:       "List widgets",
		Description: "List the widgets on a dashboard.",
	}, svc.listWidgets)

	server.Register(s, server.ToolDef{
		Name:        "dashboards_create_dashboard",
		Title:       "Create dashboard",
		Description: "Create a new dashboard for a team within a project.",
		Write:       true,
	}, svc.createDashboard)

	server.Register(s, server.ToolDef{
		Name:        "dashboards_delete_dashboard",
		Title:       "Delete dashboard",
		Description: "Delete a dashboard by its ID for a team within a project.",
		Write:       true,
		Destructive: true,
		Idempotent:  true,
	}, svc.deleteDashboard)
}

// ListDashboardsInput selects the team whose dashboards are listed.
type ListDashboardsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
}

func (s *service) listDashboards(ctx context.Context, _ *mcp.CallToolRequest, in ListDashboardsInput) (*mcp.CallToolResult, server.ListResult[Dashboard], error) {
	out, err := s.ListDashboards(ctx, in.Project, in.Team)
	return nil, server.List(out), err
}

// GetDashboardInput identifies a single dashboard.
type GetDashboardInput struct {
	Project     string `json:"project" jsonschema:"project name or ID"`
	Team        string `json:"team" jsonschema:"team name or ID"`
	DashboardID string `json:"dashboardId" jsonschema:"the dashboard ID"`
}

func (s *service) getDashboard(ctx context.Context, _ *mcp.CallToolRequest, in GetDashboardInput) (*mcp.CallToolResult, *Dashboard, error) {
	out, err := s.GetDashboard(ctx, in.Project, in.Team, in.DashboardID)
	return nil, out, err
}

// ListWidgetsInput identifies the dashboard whose widgets are listed.
type ListWidgetsInput struct {
	Project     string `json:"project" jsonschema:"project name or ID"`
	Team        string `json:"team" jsonschema:"team name or ID"`
	DashboardID string `json:"dashboardId" jsonschema:"the dashboard ID"`
}

func (s *service) listWidgets(ctx context.Context, _ *mcp.CallToolRequest, in ListWidgetsInput) (*mcp.CallToolResult, server.ListResult[Widget], error) {
	out, err := s.ListWidgets(ctx, in.Project, in.Team, in.DashboardID)
	return nil, server.List(out), err
}

// CreateDashboardInput names the dashboard to create for a team.
type CreateDashboardInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
	Name    string `json:"name" jsonschema:"the name of the dashboard"`
}

func (s *service) createDashboard(ctx context.Context, _ *mcp.CallToolRequest, in CreateDashboardInput) (*mcp.CallToolResult, *Dashboard, error) {
	out, err := s.CreateDashboard(ctx, in.Project, in.Team, in.Name)
	return nil, out, err
}

// DeleteDashboardInput identifies the dashboard to delete.
type DeleteDashboardInput struct {
	Project     string `json:"project" jsonschema:"project name or ID"`
	Team        string `json:"team" jsonschema:"team name or ID"`
	DashboardID string `json:"dashboardId" jsonschema:"the dashboard ID"`
}

func (s *service) deleteDashboard(ctx context.Context, _ *mcp.CallToolRequest, in DeleteDashboardInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.DeleteDashboard(ctx, in.Project, in.Team, in.DashboardID); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}
