// SPDX-License-Identifier: MIT

package audit

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the audit tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "audit_query_log", Title: "Query audit log",
		Description: "Query the organization-level Azure DevOps audit log. Returns a single page of decorated audit log entries; more pages may exist."}, svc.queryLog)
}

// --- Tool input types ---

// QueryLogInput filters the audit log query.
type QueryLogInput struct {
	StartTime string `json:"startTime,omitempty" jsonschema:"start of the time range as an ISO-8601 timestamp (optional)"`
	EndTime   string `json:"endTime,omitempty" jsonschema:"end of the time range as an ISO-8601 timestamp (optional)"`
	BatchSize int    `json:"batchSize,omitempty" jsonschema:"maximum number of entries to return in the page (optional)"`
}

// --- Tool handlers ---

func (s *service) queryLog(ctx context.Context, _ *mcp.CallToolRequest, in QueryLogInput) (*mcp.CallToolResult, server.ListResult[AuditLogEntry], error) {
	out, err := s.QueryLog(ctx, in.StartTime, in.EndTime, in.BatchSize)
	return nil, server.List(out), err
}
