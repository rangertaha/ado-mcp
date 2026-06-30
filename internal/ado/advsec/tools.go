// SPDX-License-Identifier: MIT

package advsec

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the Advanced Security tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "advsec_list_alerts", Title: "List security alerts",
		Description: "List Azure DevOps Advanced Security alerts for a repository, optionally filtered by alert type (code, secret or dependency)."}, svc.listAlerts)

	server.Register(s, server.ToolDef{Name: "advsec_get_alert", Title: "Get security alert",
		Description: "Get a single Azure DevOps Advanced Security alert by its numeric ID."}, svc.getAlert)
}

// --- Tool input types ---

// ListAlertsInput selects the repository and optional filters for listing
// Advanced Security alerts.
type ListAlertsInput struct {
	Project    string `json:"project" jsonschema:"the project name or ID"`
	Repository string `json:"repository" jsonschema:"the repository name or ID"`
	AlertType  string `json:"alertType,omitempty" jsonschema:"filter by alert type: code, secret or dependency (optional)"`
	Top        int    `json:"top,omitempty" jsonschema:"maximum number of alerts to return (optional)"`
}

// GetAlertInput identifies a single Advanced Security alert.
type GetAlertInput struct {
	Project    string `json:"project" jsonschema:"the project name or ID"`
	Repository string `json:"repository" jsonschema:"the repository name or ID"`
	AlertID    int    `json:"alertId" jsonschema:"the numeric alert ID"`
}

// --- Tool handlers ---

func (s *service) listAlerts(ctx context.Context, _ *mcp.CallToolRequest, in ListAlertsInput) (*mcp.CallToolResult, server.ListResult[Alert], error) {
	out, err := s.ListAlerts(ctx, in.Project, in.Repository, in.AlertType, in.Top)
	return nil, server.List(out), err
}

func (s *service) getAlert(ctx context.Context, _ *mcp.CallToolRequest, in GetAlertInput) (*mcp.CallToolResult, *Alert, error) {
	out, err := s.GetAlert(ctx, in.Project, in.Repository, in.AlertID)
	return nil, out, err
}
