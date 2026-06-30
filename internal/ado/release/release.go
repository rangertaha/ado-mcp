// SPDX-License-Identifier: MIT

// Package release provides Azure DevOps Release Management operations:
// release definitions, releases, deployments, and approvals.
package release

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name for this service area.
const Name = "release"

// service holds the Azure DevOps host clients used by this area.
type service struct{ c *ado.Clients }

// ReleaseDefinition describes a release pipeline definition.
type ReleaseDefinition struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Path string `json:"path,omitempty"`
	URL  string `json:"url,omitempty"`
}

// Release describes a single release instance.
type Release struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status,omitempty"`
	CreatedOn string `json:"createdOn,omitempty"`
	URL       string `json:"url,omitempty"`
}

// Deployment describes a deployment of a release to an environment.
type Deployment struct {
	ID                int    `json:"id"`
	DeploymentStatus  string `json:"deploymentStatus,omitempty"`
	Release           any    `json:"release,omitempty"`
	ReleaseDefinition any    `json:"releaseDefinition,omitempty"`
}

// Approval describes a release approval request.
type Approval struct {
	ID       int    `json:"id"`
	Status   string `json:"status,omitempty"`
	Approver any    `json:"approver,omitempty"`
	Release  any    `json:"release,omitempty"`
}

// ManualIntervention describes a manual intervention awaiting action in a release.
type ManualIntervention struct {
	ID       int    `json:"id"`
	Name     string `json:"name,omitempty"`
	Status   string `json:"status,omitempty"`
	Comments string `json:"comments,omitempty"`
	Release  any    `json:"release,omitempty"`
}

// ReleaseEnvironment describes an environment within a release.
type ReleaseEnvironment struct {
	ID     int    `json:"id"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
}

// ListManualInterventions lists the manual interventions for a release.
func (s *service) ListManualInterventions(ctx context.Context, project string, releaseID int) ([]ManualIntervention, error) {
	var out client.List[ManualIntervention]
	path := fmt.Sprintf("/%s/_apis/release/releases/%d/manualinterventions", url.PathEscape(project), releaseID)
	if err := s.c.VSRM.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// UpdateManualIntervention approves or rejects a manual intervention in a release.
func (s *service) UpdateManualIntervention(ctx context.Context, project string, releaseID, interventionID int, status, comment string) (*ManualIntervention, error) {
	body := map[string]any{"status": status}
	if comment != "" {
		body["comment"] = comment
	}
	var out ManualIntervention
	path := fmt.Sprintf("/%s/_apis/release/releases/%d/manualinterventions/%d", url.PathEscape(project), releaseID, interventionID)
	if err := s.c.VSRM.PatchJSON(ctx, path, nil, body, &out, ""); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeployEnvironment starts a deployment of a release to an environment.
func (s *service) DeployEnvironment(ctx context.Context, project string, releaseID, environmentID int) (*ReleaseEnvironment, error) {
	body := map[string]any{"status": "inProgress"}
	var out ReleaseEnvironment
	path := fmt.Sprintf("/%s/_apis/release/releases/%d/environments/%d", url.PathEscape(project), releaseID, environmentID)
	if err := s.c.VSRM.PatchJSON(ctx, path, nil, body, &out, ""); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDefinitions lists the release definitions in a project.
func (s *service) ListDefinitions(ctx context.Context, project string) ([]ReleaseDefinition, error) {
	var out client.List[ReleaseDefinition]
	path := fmt.Sprintf("/%s/_apis/release/definitions", url.PathEscape(project))
	if err := s.c.VSRM.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListReleases lists releases in a project, optionally filtered by definition ID.
func (s *service) ListReleases(ctx context.Context, project string, definitionID, top int) ([]Release, error) {
	q := url.Values{}
	if definitionID > 0 {
		q.Set("definitionId", strconv.Itoa(definitionID))
	}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[Release]
	path := fmt.Sprintf("/%s/_apis/release/releases", url.PathEscape(project))
	if err := s.c.VSRM.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetRelease gets a single release by ID.
func (s *service) GetRelease(ctx context.Context, project string, releaseID int) (*Release, error) {
	var out Release
	path := fmt.Sprintf("/%s/_apis/release/releases/%d", url.PathEscape(project), releaseID)
	if err := s.c.VSRM.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateRelease creates a new release from a release definition.
func (s *service) CreateRelease(ctx context.Context, project string, definitionID int, description string) (*Release, error) {
	body := map[string]any{"definitionId": definitionID}
	if description != "" {
		body["description"] = description
	}
	var out Release
	path := fmt.Sprintf("/%s/_apis/release/releases", url.PathEscape(project))
	if err := s.c.VSRM.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListDeployments lists deployments in a project, optionally filtered by definition ID.
func (s *service) ListDeployments(ctx context.Context, project string, definitionID, top int) ([]Deployment, error) {
	q := url.Values{}
	if definitionID > 0 {
		q.Set("definitionId", strconv.Itoa(definitionID))
	}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[Deployment]
	path := fmt.Sprintf("/%s/_apis/release/deployments", url.PathEscape(project))
	if err := s.c.VSRM.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListApprovals lists approvals in a project, optionally filtered by status.
func (s *service) ListApprovals(ctx context.Context, project, statusFilter string) ([]Approval, error) {
	q := url.Values{}
	if statusFilter != "" {
		q.Set("statusFilter", statusFilter)
	}
	var out client.List[Approval]
	path := fmt.Sprintf("/%s/_apis/release/approvals", url.PathEscape(project))
	if err := s.c.VSRM.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// UpdateApproval approves or rejects a release approval.
func (s *service) UpdateApproval(ctx context.Context, project string, approvalID int, status, comments string) (*Approval, error) {
	body := map[string]any{"status": status}
	if comments != "" {
		body["comments"] = comments
	}
	var out Approval
	path := fmt.Sprintf("/%s/_apis/release/approvals/%d", url.PathEscape(project), approvalID)
	if err := s.c.VSRM.PatchJSON(ctx, path, nil, body, &out, ""); err != nil {
		return nil, err
	}
	return &out, nil
}
