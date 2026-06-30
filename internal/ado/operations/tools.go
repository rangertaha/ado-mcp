// SPDX-License-Identifier: MIT

package operations

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the operations tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "operations_get", Title: "Get operation",
		Description: "Get the status of a long-running asynchronous Azure DevOps operation by its operation ID."}, svc.get)
}

// --- Tool input types ---

// GetInput identifies the operation to retrieve.
type GetInput struct {
	OperationID string `json:"operationId" jsonschema:"the ID of the operation to retrieve"`
}

// --- Tool handlers ---

func (s *service) get(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, *Operation, error) {
	out, err := s.Get(ctx, in.OperationID)
	return nil, out, err
}
