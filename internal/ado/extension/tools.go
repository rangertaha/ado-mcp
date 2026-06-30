// SPDX-License-Identifier: MIT

package extension

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the extension management tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "extension_list_installed", Title: "List installed extensions",
		Description: "List all extensions installed in the Azure DevOps organization."}, svc.listInstalled)

	server.Register(s, server.ToolDef{Name: "extension_get", Title: "Get installed extension",
		Description: "Get a single installed extension by its publisher and extension identifiers."}, svc.get)
}

// --- Tool input types ---

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// GetInput identifies an installed extension to retrieve.
type GetInput struct {
	PublisherID string `json:"publisherId" jsonschema:"identifier of the extension publisher"`
	ExtensionID string `json:"extensionId" jsonschema:"identifier of the extension"`
}

// --- Tool handlers ---

func (s *service) listInstalled(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[InstalledExtension], error) {
	out, err := s.ListInstalled(ctx)
	return nil, server.List(out), err
}

func (s *service) get(ctx context.Context, _ *mcp.CallToolRequest, in GetInput) (*mcp.CallToolResult, *InstalledExtension, error) {
	out, err := s.Get(ctx, in.PublisherID, in.ExtensionID)
	return nil, out, err
}
