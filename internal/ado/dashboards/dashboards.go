// SPDX-License-Identifier: MIT

// Package dashboards exposes the Azure DevOps Dashboards service:
// team-scoped dashboards and the widgets they contain, including
// creating and deleting dashboards.
package dashboards

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "dashboards"

// service wraps the Azure DevOps clients for Dashboards operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Dashboard is a team dashboard composed of widgets.
type Dashboard struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	OwnerID     string `json:"ownerId,omitempty"`
	Position    int    `json:"position,omitempty"`
	Widgets     any    `json:"widgets,omitempty"`
	URL         string `json:"url,omitempty"`
}

// Widget is a single widget placed on a dashboard.
type Widget struct {
	ID             string `json:"id"`
	Name           string `json:"name,omitempty"`
	Position       any    `json:"position,omitempty"`
	Size           any    `json:"size,omitempty"`
	ContributionID string `json:"contributionId,omitempty"`
	URL            string `json:"url,omitempty"`
}

// --- Operations ---

// ListDashboards returns the dashboards for a team within a project.
func (s *service) ListDashboards(ctx context.Context, project, team string) ([]Dashboard, error) {
	var out client.List[Dashboard]
	path := fmt.Sprintf("/%s/%s/_apis/dashboard/dashboards",
		url.PathEscape(project), url.PathEscape(team))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// CreateDashboard creates a new dashboard for a team within a project.
func (s *service) CreateDashboard(ctx context.Context, project, team, name string) (*Dashboard, error) {
	var out Dashboard
	path := fmt.Sprintf("/%s/%s/_apis/dashboard/dashboards",
		url.PathEscape(project), url.PathEscape(team))
	body := map[string]string{"name": name}
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteDashboard deletes a dashboard by its ID for a team within a project.
func (s *service) DeleteDashboard(ctx context.Context, project, team, dashboardID string) error {
	path := fmt.Sprintf("/%s/%s/_apis/dashboard/dashboards/%s",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(dashboardID))
	return s.c.Org.Delete(ctx, path, nil, nil)
}

// GetDashboard returns a single dashboard by its ID.
func (s *service) GetDashboard(ctx context.Context, project, team, dashboardID string) (*Dashboard, error) {
	var out Dashboard
	path := fmt.Sprintf("/%s/%s/_apis/dashboard/dashboards/%s",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(dashboardID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListWidgets returns the widgets on a dashboard.
func (s *service) ListWidgets(ctx context.Context, project, team, dashboardID string) ([]Widget, error) {
	var out client.List[Widget]
	path := fmt.Sprintf("/%s/%s/_apis/dashboard/dashboards/%s/widgets",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(dashboardID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
