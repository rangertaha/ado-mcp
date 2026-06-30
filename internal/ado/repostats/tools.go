// SPDX-License-Identifier: MIT

package repostats

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the statistics/survey toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "stats_repo_contributors", Title: "Repo contributors",
		Description: "Contribution stats (commits and change counts per author) for one repository. Optionally filter by date range."}, svc.repoContributors)
	server.Register(s, server.ToolDef{Name: "stats_project_contributors", Title: "Project contributors",
		Description: "Contribution stats aggregated across every repository in a project."}, svc.projectContributors)
	server.Register(s, server.ToolDef{Name: "stats_org_contributors", Title: "Org contributors",
		Description: "Contribution stats aggregated across every repository in every project in the organization. Can be slow on large orgs."}, svc.orgContributors)
	server.Register(s, server.ToolDef{Name: "stats_pull_requests", Title: "Pull request stats",
		Description: "Pull-request activity per author and by status, for a repo or for a whole project (omit repo)."}, svc.pullRequestStats)
	server.Register(s, server.ToolDef{Name: "stats_work_items", Title: "Work item stats",
		Description: "Work item breakdown by type, state, and assignee for a project (optionally scoped by a WIQL query)."}, svc.workItemStats)
	server.Register(s, server.ToolDef{Name: "stats_builds", Title: "Build stats",
		Description: "Build outcomes (success/failure rates) per definition for a project."}, svc.buildStats)
}

// --- Inputs ---

// RepoContributorsInput scopes a repository contribution survey.
type RepoContributorsInput struct {
	Project    string `json:"project" jsonschema:"project name or ID"`
	Repo       string `json:"repo" jsonschema:"repository name or ID"`
	FromDate   string `json:"fromDate,omitempty" jsonschema:"only count commits on/after this date (optional), e.g. 2026-01-01"`
	ToDate     string `json:"toDate,omitempty" jsonschema:"only count commits on/before this date (optional)"`
	MaxCommits int    `json:"maxCommits,omitempty" jsonschema:"max commits to scan (default 2000)"`
}

// ProjectStatsInput scopes a project-wide survey.
type ProjectStatsInput struct {
	Project    string `json:"project" jsonschema:"project name or ID"`
	FromDate   string `json:"fromDate,omitempty" jsonschema:"only count items on/after this date (optional)"`
	ToDate     string `json:"toDate,omitempty" jsonschema:"only count items on/before this date (optional)"`
	MaxCommits int    `json:"maxCommits,omitempty" jsonschema:"max commits to scan per repository (default 2000)"`
}

// OrgStatsInput scopes an organization-wide survey.
type OrgStatsInput struct {
	FromDate   string `json:"fromDate,omitempty" jsonschema:"only count commits on/after this date (optional)"`
	ToDate     string `json:"toDate,omitempty" jsonschema:"only count commits on/before this date (optional)"`
	MaxCommits int    `json:"maxCommits,omitempty" jsonschema:"max commits to scan per repository (default 2000)"`
}

// PullRequestStatsInput scopes a PR survey.
type PullRequestStatsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Repo    string `json:"repo,omitempty" jsonschema:"repository name or ID; omit to survey all repos in the project"`
	Status  string `json:"status,omitempty" jsonschema:"status filter: all (default), active, completed, abandoned"`
	MaxPRs  int    `json:"maxPRs,omitempty" jsonschema:"max pull requests to scan per repository (default 2000)"`
}

// WorkItemStatsInput scopes a work-item survey.
type WorkItemStatsInput struct {
	Project  string `json:"project" jsonschema:"project name or ID"`
	Wiql     string `json:"wiql,omitempty" jsonschema:"optional WIQL query to scope the items (defaults to all work items in the project)"`
	MaxItems int    `json:"maxItems,omitempty" jsonschema:"max work items to scan (default 2000)"`
}

// BuildStatsInput scopes a build survey.
type BuildStatsInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	DefinitionID int    `json:"definitionId,omitempty" jsonschema:"restrict to a single build definition (optional)"`
	MaxBuilds    int    `json:"maxBuilds,omitempty" jsonschema:"max builds to scan (default 2000)"`
}

// --- Handlers ---

func (s *service) repoContributors(ctx context.Context, _ *mcp.CallToolRequest, in RepoContributorsInput) (*mcp.CallToolResult, *Stats, error) {
	out, err := s.RepoStats(ctx, in.Project, in.Repo, in.FromDate, in.ToDate, in.MaxCommits)
	return nil, out, err
}

func (s *service) projectContributors(ctx context.Context, _ *mcp.CallToolRequest, in ProjectStatsInput) (*mcp.CallToolResult, *Stats, error) {
	out, err := s.ProjectStats(ctx, in.Project, in.FromDate, in.ToDate, in.MaxCommits)
	return nil, out, err
}

func (s *service) orgContributors(ctx context.Context, _ *mcp.CallToolRequest, in OrgStatsInput) (*mcp.CallToolResult, *Stats, error) {
	out, err := s.OrgStats(ctx, in.FromDate, in.ToDate, in.MaxCommits)
	return nil, out, err
}

func (s *service) pullRequestStats(ctx context.Context, _ *mcp.CallToolRequest, in PullRequestStatsInput) (*mcp.CallToolResult, *PRStats, error) {
	out, err := s.PullRequestStats(ctx, in.Project, in.Repo, in.Status, in.MaxPRs)
	return nil, out, err
}

func (s *service) workItemStats(ctx context.Context, _ *mcp.CallToolRequest, in WorkItemStatsInput) (*mcp.CallToolResult, *WorkItemStats, error) {
	out, err := s.WorkItemStatsForQuery(ctx, in.Project, in.Wiql, in.MaxItems)
	return nil, out, err
}

func (s *service) buildStats(ctx context.Context, _ *mcp.CallToolRequest, in BuildStatsInput) (*mcp.CallToolResult, *BuildStats, error) {
	out, err := s.BuildStatsForProject(ctx, in.Project, in.DefinitionID, in.MaxBuilds)
	return nil, out, err
}
