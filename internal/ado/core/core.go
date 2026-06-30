// SPDX-License-Identifier: MIT

// Package core exposes the Azure DevOps Core service: projects, teams, team
// members and processes.
package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "core"

// service wraps the Azure DevOps clients for Core operations.
type service struct {
	c *ado.Clients
}

// --- Domain types (trimmed to the fields useful to an LLM client) ---

// Project is an Azure DevOps team project.
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	State       string `json:"state,omitempty"`
	Revision    int    `json:"revision,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	LastUpdate  string `json:"lastUpdateTime,omitempty"`
}

// Team is a team within a project.
type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// Member is a member of a team.
type Member struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	UniqueName  string `json:"uniqueName,omitempty"`
	Descriptor  string `json:"descriptor,omitempty"`
}

// teamMember is the wire shape returned by the team members endpoint.
type teamMember struct {
	Identity Member `json:"identity"`
}

// Process is an inherited/process template.
type Process struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// --- Operations ---

// ListProjects returns the projects in the organization.
func (s *service) ListProjects(ctx context.Context, top int, skip int) ([]Project, error) {
	q := url.Values{}
	if top > 0 {
		q.Set("$top", fmt.Sprintf("%d", top))
	}
	if skip > 0 {
		q.Set("$skip", fmt.Sprintf("%d", skip))
	}
	var out client.List[Project]
	if err := s.c.Org.GetJSON(ctx, "/_apis/projects", q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetProject returns a single project by name or ID.
func (s *service) GetProject(ctx context.Context, nameOrID string) (*Project, error) {
	var p Project
	path := fmt.Sprintf("/_apis/projects/%s", url.PathEscape(nameOrID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListTeams returns the teams in a project.
func (s *service) ListTeams(ctx context.Context, project string) ([]Team, error) {
	var out client.List[Team]
	path := fmt.Sprintf("/_apis/projects/%s/teams", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetTeam returns a single team within a project by name or ID.
func (s *service) GetTeam(ctx context.Context, project, team string) (*Team, error) {
	var t Team
	path := fmt.Sprintf("/_apis/projects/%s/teams/%s", url.PathEscape(project), url.PathEscape(team))
	if err := s.c.Org.GetJSON(ctx, path, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// ListTeamMembers returns the members of a team.
func (s *service) ListTeamMembers(ctx context.Context, project, team string) ([]Member, error) {
	var out client.List[teamMember]
	path := fmt.Sprintf("/_apis/projects/%s/teams/%s/members", url.PathEscape(project), url.PathEscape(team))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	members := make([]Member, 0, len(out.Value))
	for _, m := range out.Value {
		members = append(members, m.Identity)
	}
	return members, nil
}

// ListProcesses returns the process templates available to the organization.
func (s *service) ListProcesses(ctx context.Context) ([]Process, error) {
	var out client.List[Process]
	if err := s.c.Org.GetJSON(ctx, "/_apis/process/processes", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
