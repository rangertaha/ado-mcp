// SPDX-License-Identifier: MIT

// Package approvals exposes Azure DevOps pipeline run approvals and checks:
// retrieving approvals, listing check configurations for protected resources,
// and approving or rejecting pending approvals.
package approvals

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "approvals"

// apiVersion is the preview API version required by the approvals and checks
// endpoints.
const apiVersion = "7.1-preview.1"

// service wraps the Azure DevOps clients for approvals operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Approval is a pipeline run approval, including its current status and the
// approval steps.
type Approval struct {
	ID                   string `json:"id"`
	Status               string `json:"status,omitempty"`
	Steps                any    `json:"steps,omitempty"`
	BlockedApprovers     any    `json:"blockedApprovers,omitempty"`
	Instructions         string `json:"instructions,omitempty"`
	MinRequiredApprovers int    `json:"minRequiredApprovers,omitempty"`
}

// CheckConfiguration is a check configured on a protected resource (such as a
// service connection, environment, or variable group).
type CheckConfiguration struct {
	ID       int `json:"id"`
	Type     any `json:"type,omitempty"`
	Resource any `json:"resource,omitempty"`
	Settings any `json:"settings,omitempty"`
}

// approvalUpdate is one element of the array body sent to update approvals.
type approvalUpdate struct {
	ApprovalID string `json:"approvalId"`
	Status     string `json:"status"`
	Comment    string `json:"comment,omitempty"`
}

// --- Approvals operations ---

// Get retrieves a single approval by its identifier.
func (s *service) Get(ctx context.Context, project, approvalID string) (*Approval, error) {
	path := fmt.Sprintf("/%s/_apis/pipelines/approvals/%s",
		url.PathEscape(project), url.PathEscape(approvalID))
	var out Approval
	if err := s.c.Org.GetJSONVersion(ctx, path, nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListCheckConfigurations lists the check configurations defined on the
// protected resource identified by resourceType and resourceID.
func (s *service) ListCheckConfigurations(ctx context.Context, project, resourceType, resourceID string) ([]CheckConfiguration, error) {
	path := fmt.Sprintf("/%s/_apis/pipelines/checks/configurations", url.PathEscape(project))
	q := url.Values{}
	q.Set("resourceType", resourceType)
	q.Set("resourceId", resourceID)
	var out client.List[CheckConfiguration]
	if err := s.c.Org.GetJSONVersion(ctx, path, q, &out, apiVersion); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// Update approves or rejects a pending approval. status must be "approved" or
// "rejected"; comment is optional. It returns the updated approval records.
func (s *service) Update(ctx context.Context, project, approvalID, status, comment string) ([]any, error) {
	path := fmt.Sprintf("/%s/_apis/pipelines/approvals", url.PathEscape(project))
	body := []approvalUpdate{{ApprovalID: approvalID, Status: status, Comment: comment}}
	var out client.List[any]
	if _, err := s.c.Org.Do(ctx, client.Request{
		Method:     "PATCH",
		Path:       path,
		Body:       body,
		APIVersion: apiVersion,
		Out:        &out,
	}); err != nil {
		return nil, err
	}
	return out.Value, nil
}
