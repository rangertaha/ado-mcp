// SPDX-License-Identifier: MIT

package release

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Release Management toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "release_list_definitions", Title: "List release definitions",
		Description: "List the release definitions in a project."}, svc.listDefinitions)
	server.Register(s, server.ToolDef{Name: "release_list_releases", Title: "List releases",
		Description: "List releases in a project, optionally filtered by release definition."}, svc.listReleases)
	server.Register(s, server.ToolDef{Name: "release_get_release", Title: "Get release",
		Description: "Get a single release by ID."}, svc.getRelease)
	server.Register(s, server.ToolDef{Name: "release_list_deployments", Title: "List deployments",
		Description: "List deployments in a project, optionally filtered by release definition."}, svc.listDeployments)
	server.Register(s, server.ToolDef{Name: "release_list_approvals", Title: "List approvals",
		Description: "List release approvals in a project, optionally filtered by status."}, svc.listApprovals)

	// --- Write tools ---
	server.Register(s, server.ToolDef{Name: "release_create_release", Title: "Create release",
		Description: "Create a new release from a release definition.", Write: true}, svc.createRelease)
	server.Register(s, server.ToolDef{Name: "release_update_approval", Title: "Update approval",
		Description: "Approve or reject a release approval.", Write: true, Idempotent: true}, svc.updateApproval)
	server.Register(s, server.ToolDef{Name: "release_list_manual_interventions", Title: "List manual interventions",
		Description: "List the manual interventions for a release."}, svc.listManualInterventions)
	server.Register(s, server.ToolDef{Name: "release_update_manual_intervention", Title: "Update manual intervention",
		Description: "Approve or reject a manual intervention in a release.", Write: true, Idempotent: true}, svc.updateManualIntervention)
	server.Register(s, server.ToolDef{Name: "release_deploy_environment", Title: "Deploy environment",
		Description: "Start a deployment of a release to an environment.", Write: true}, svc.deployEnvironment)
}

// --- Tool input types ---

// ProjectInput identifies a project.
type ProjectInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

// ListReleasesInput filters releases.
type ListReleasesInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	DefinitionID int    `json:"definitionId,omitempty" jsonschema:"restrict to a release definition (optional)"`
	Top          int    `json:"top,omitempty" jsonschema:"maximum number of releases (optional)"`
}

// ReleaseInput identifies a release.
type ReleaseInput struct {
	Project   string `json:"project" jsonschema:"project name or ID"`
	ReleaseID int    `json:"releaseId" jsonschema:"release ID"`
}

// CreateReleaseInput creates a release.
type CreateReleaseInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	DefinitionID int    `json:"definitionId" jsonschema:"release definition ID to create the release from"`
	Description  string `json:"description,omitempty" jsonschema:"description for the new release (optional)"`
}

// ListDeploymentsInput filters deployments.
type ListDeploymentsInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	DefinitionID int    `json:"definitionId,omitempty" jsonschema:"restrict to a release definition (optional)"`
	Top          int    `json:"top,omitempty" jsonschema:"maximum number of deployments (optional)"`
}

// ListApprovalsInput filters approvals.
type ListApprovalsInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	StatusFilter string `json:"statusFilter,omitempty" jsonschema:"filter by approval status, e.g. pending (optional)"`
}

// UpdateApprovalInput approves or rejects an approval.
type UpdateApprovalInput struct {
	Project    string `json:"project" jsonschema:"project name or ID"`
	ApprovalID int    `json:"approvalId" jsonschema:"approval ID"`
	Status     string `json:"status" jsonschema:"new status: approved or rejected"`
	Comments   string `json:"comments,omitempty" jsonschema:"comments for the approval decision (optional)"`
}

// ListManualInterventionsInput identifies a release.
type ListManualInterventionsInput struct {
	Project   string `json:"project" jsonschema:"project name or ID"`
	ReleaseID int    `json:"releaseId" jsonschema:"release ID"`
}

// UpdateManualInterventionInput approves or rejects a manual intervention.
type UpdateManualInterventionInput struct {
	Project        string `json:"project" jsonschema:"project name or ID"`
	ReleaseID      int    `json:"releaseId" jsonschema:"release ID"`
	InterventionID int    `json:"interventionId" jsonschema:"manual intervention ID"`
	Status         string `json:"status" jsonschema:"new status: approved or rejected"`
	Comment        string `json:"comment,omitempty" jsonschema:"comment for the intervention decision (optional)"`
}

// DeployEnvironmentInput identifies a release environment to deploy.
type DeployEnvironmentInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	ReleaseID     int    `json:"releaseId" jsonschema:"release ID"`
	EnvironmentID int    `json:"environmentId" jsonschema:"release environment ID"`
}

// --- Tool handlers ---

func (s *service) listDefinitions(ctx context.Context, _ *mcp.CallToolRequest, in ProjectInput) (*mcp.CallToolResult, server.ListResult[ReleaseDefinition], error) {
	out, err := s.ListDefinitions(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) listReleases(ctx context.Context, _ *mcp.CallToolRequest, in ListReleasesInput) (*mcp.CallToolResult, server.ListResult[Release], error) {
	out, err := s.ListReleases(ctx, in.Project, in.DefinitionID, in.Top)
	return nil, server.List(out), err
}

func (s *service) getRelease(ctx context.Context, _ *mcp.CallToolRequest, in ReleaseInput) (*mcp.CallToolResult, *Release, error) {
	out, err := s.GetRelease(ctx, in.Project, in.ReleaseID)
	return nil, out, err
}

func (s *service) listDeployments(ctx context.Context, _ *mcp.CallToolRequest, in ListDeploymentsInput) (*mcp.CallToolResult, server.ListResult[Deployment], error) {
	out, err := s.ListDeployments(ctx, in.Project, in.DefinitionID, in.Top)
	return nil, server.List(out), err
}

func (s *service) listApprovals(ctx context.Context, _ *mcp.CallToolRequest, in ListApprovalsInput) (*mcp.CallToolResult, server.ListResult[Approval], error) {
	out, err := s.ListApprovals(ctx, in.Project, in.StatusFilter)
	return nil, server.List(out), err
}

func (s *service) createRelease(ctx context.Context, _ *mcp.CallToolRequest, in CreateReleaseInput) (*mcp.CallToolResult, *Release, error) {
	out, err := s.CreateRelease(ctx, in.Project, in.DefinitionID, in.Description)
	return nil, out, err
}

func (s *service) updateApproval(ctx context.Context, _ *mcp.CallToolRequest, in UpdateApprovalInput) (*mcp.CallToolResult, *Approval, error) {
	out, err := s.UpdateApproval(ctx, in.Project, in.ApprovalID, in.Status, in.Comments)
	return nil, out, err
}

func (s *service) listManualInterventions(ctx context.Context, _ *mcp.CallToolRequest, in ListManualInterventionsInput) (*mcp.CallToolResult, server.ListResult[ManualIntervention], error) {
	out, err := s.ListManualInterventions(ctx, in.Project, in.ReleaseID)
	return nil, server.List(out), err
}

func (s *service) updateManualIntervention(ctx context.Context, _ *mcp.CallToolRequest, in UpdateManualInterventionInput) (*mcp.CallToolResult, *ManualIntervention, error) {
	out, err := s.UpdateManualIntervention(ctx, in.Project, in.ReleaseID, in.InterventionID, in.Status, in.Comment)
	return nil, out, err
}

func (s *service) deployEnvironment(ctx context.Context, _ *mcp.CallToolRequest, in DeployEnvironmentInput) (*mcp.CallToolResult, *ReleaseEnvironment, error) {
	out, err := s.DeployEnvironment(ctx, in.Project, in.ReleaseID, in.EnvironmentID)
	return nil, out, err
}
