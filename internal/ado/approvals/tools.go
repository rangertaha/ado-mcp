// SPDX-License-Identifier: MIT

package approvals

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the approvals tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "approvals_get", Title: "Get approval",
		Description: "Get a pipeline run approval by its identifier."}, svc.get)

	server.Register(s, server.ToolDef{Name: "approvals_list_check_configurations", Title: "List check configurations",
		Description: "List the check configurations defined on a protected resource (by resource type and id)."}, svc.listCheckConfigurations)

	server.Register(s, server.ToolDef{Name: "approvals_update", Title: "Update approval",
		Description: "Approve or reject a pending pipeline run approval.",
		Write:       true, Idempotent: true}, svc.update)
}

// --- Tool input types ---

// GetInput identifies the approval to retrieve.
type GetInput struct {
	Project    string `json:"project" jsonschema:"project ID or name"`
	ApprovalID string `json:"approvalId" jsonschema:"the approval identifier"`
}

// ListCheckConfigurationsInput identifies the protected resource whose check
// configurations are listed.
type ListCheckConfigurationsInput struct {
	Project      string `json:"project" jsonschema:"project ID or name"`
	ResourceType string `json:"resourceType" jsonschema:"the protected resource type (for example endpoint, environment, or variablegroup)"`
	ResourceID   string `json:"resourceId" jsonschema:"the protected resource identifier"`
}

// UpdateInput approves or rejects an approval.
type UpdateInput struct {
	Project    string `json:"project" jsonschema:"project ID or name"`
	ApprovalID string `json:"approvalId" jsonschema:"the approval identifier"`
	Status     string `json:"status" jsonschema:"the new approval status: approved or rejected"`
	Comment    string `json:"comment,omitempty" jsonschema:"an optional comment to record with the decision"`
}

// --- Tool handlers ---

func (s *service) get(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, *Approval, error) {
	out, err := s.Get(ctx, in.Project, in.ApprovalID)
	return nil, out, err
}

func (s *service) listCheckConfigurations(ctx context.Context, _ *mcp.CallToolRequest, in ListCheckConfigurationsInput) (*mcp.CallToolResult, server.ListResult[CheckConfiguration], error) {
	out, err := s.ListCheckConfigurations(ctx, in.Project, in.ResourceType, in.ResourceID)
	return nil, server.List(out), err
}

func (s *service) update(ctx context.Context, _ *mcp.CallToolRequest, in UpdateInput) (*mcp.CallToolResult, server.ListResult[any], error) {
	out, err := s.Update(ctx, in.Project, in.ApprovalID, in.Status, in.Comment)
	return nil, server.List(out), err
}
