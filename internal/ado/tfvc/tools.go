// SPDX-License-Identifier: MIT

package tfvc

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the TFVC toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "tfvc_list_changesets", Title: "List changesets",
		Description: "List TFVC changesets, optionally filtered by item path."}, svc.listChangesets)
	server.Register(s, server.ToolDef{Name: "tfvc_get_changeset", Title: "Get changeset",
		Description: "Get a single TFVC changeset by ID."}, svc.getChangeset)
	server.Register(s, server.ToolDef{Name: "tfvc_list_branches", Title: "List branches",
		Description: "List the TFVC branches, including child branches."}, svc.listBranches)
	server.Register(s, server.ToolDef{Name: "tfvc_get_item", Title: "Get item",
		Description: "Get a TFVC item at the given path, including its content."}, svc.getItem)
}

// --- Tool input types ---

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// ListChangesetsInput filters and limits the changesets returned.
type ListChangesetsInput struct {
	ItemPath string `json:"itemPath,omitempty" jsonschema:"filter changesets by item path, e.g. $/Project/Folder (optional)"`
	Top      int    `json:"top,omitempty" jsonschema:"maximum number of changesets to return (optional)"`
}

// GetChangesetInput identifies a changeset.
type GetChangesetInput struct {
	ID int `json:"id" jsonschema:"the changeset ID"`
}

// GetItemInput identifies a TFVC item to read.
type GetItemInput struct {
	Path string `json:"path" jsonschema:"the TFVC item path, e.g. $/Project/Folder/file.txt"`
}

// --- Tool handlers ---

func (s *service) listChangesets(ctx context.Context, _ *mcp.CallToolRequest, in ListChangesetsInput) (*mcp.CallToolResult, server.ListResult[Changeset], error) {
	out, err := s.ListChangesets(ctx, in.ItemPath, in.Top)
	return nil, server.List(out), err
}

func (s *service) getChangeset(ctx context.Context, _ *mcp.CallToolRequest, in GetChangesetInput) (*mcp.CallToolResult, *Changeset, error) {
	out, err := s.GetChangeset(ctx, in.ID)
	return nil, out, err
}

func (s *service) listBranches(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[TfvcBranch], error) {
	out, err := s.ListBranches(ctx)
	return nil, server.List(out), err
}

func (s *service) getItem(ctx context.Context, _ *mcp.CallToolRequest, in GetItemInput) (*mcp.CallToolResult, *TfvcItem, error) {
	out, err := s.GetItem(ctx, in.Path)
	return nil, out, err
}
