// SPDX-License-Identifier: MIT

package wiki

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Wiki toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "wiki_list_wikis", Title: "List wikis",
		Description: "List the wikis in a project."}, svc.listWikis)
	server.Register(s, server.ToolDef{Name: "wiki_get_wiki", Title: "Get wiki",
		Description: "Get a single wiki by ID or name."}, svc.getWiki)
	server.Register(s, server.ToolDef{Name: "wiki_get_page", Title: "Get wiki page",
		Description: "Get a wiki page by path, optionally including its content and (with recursionLevel='full') its subpages."}, svc.getPage)

	// --- Write tools ---
	server.Register(s, server.ToolDef{Name: "wiki_create_or_update_page", Title: "Create or update wiki page",
		Description: "Create a new wiki page or update an existing one at the given path. Updating an existing page normally requires an If-Match ETag; this tool issues the PUT directly and the API may report a conflict if the page already exists.",
		Write:       true, Idempotent: true}, svc.createOrUpdatePage)
	server.Register(s, server.ToolDef{Name: "wiki_create_wiki", Title: "Create wiki",
		Description: "Create a new project or code wiki in a project. type must be 'projectWiki' or 'codeWiki'.",
		Write:       true}, svc.createWiki)
	server.Register(s, server.ToolDef{Name: "wiki_delete_page", Title: "Delete wiki page",
		Description: "Delete a wiki page at the given path.",
		Write:       true, Destructive: true, Idempotent: true}, svc.deletePage)
	server.Register(s, server.ToolDef{Name: "wiki_create_attachment", Title: "Create wiki attachment",
		Description: "Upload an attachment to a wiki. Binary content must be base64-encoded.",
		Write:       true}, svc.createAttachment)
	server.Register(s, server.ToolDef{Name: "wiki_move_page", Title: "Move wiki page",
		Description: "Move or rename a wiki page from one path to another.",
		Write:       true, Idempotent: true}, svc.movePage)
	server.Register(s, server.ToolDef{Name: "wiki_get_page_by_id", Title: "Get wiki page by ID",
		Description: "Get a wiki page by its numeric page ID, optionally including content."}, svc.getPageByID)
	server.Register(s, server.ToolDef{Name: "wiki_get_page_stats", Title: "Get wiki page stats",
		Description: "Get view statistics for a wiki page over the last N days."}, svc.getPageStats)
	server.Register(s, server.ToolDef{Name: "wiki_update_wiki", Title: "Rename wiki",
		Description: "Rename a wiki.", Write: true, Idempotent: true}, svc.updateWiki)
	server.Register(s, server.ToolDef{Name: "wiki_delete_wiki", Title: "Delete wiki",
		Description: "Delete (unpublish) a wiki by ID or name.",
		Write:       true, Destructive: true, Idempotent: true}, svc.deleteWiki)
	server.Register(s, server.ToolDef{Name: "wiki_append_to_page", Title: "Append to wiki page",
		Description: "Append content to an existing wiki page (creating it if absent) without resending the whole page. Useful for adding changelog entries, notes, or sections.",
		Write:       true}, svc.appendToPage)
	server.Register(s, server.ToolDef{Name: "wiki_list_pages", Title: "List wiki pages",
		Description: "List the page tree of a wiki (expanded to full depth) from a root path, for browsing structure."}, svc.listPages)
}

// --- Tool input types ---

// ListWikisInput identifies a project.
type ListWikisInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

// GetWikiInput identifies a wiki.
type GetWikiInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
}

// GetPageInput identifies a wiki page to read.
type GetPageInput struct {
	Project        string `json:"project" jsonschema:"project name or ID"`
	WikiID         string `json:"wikiId" jsonschema:"wiki ID or name"`
	Path           string `json:"path" jsonschema:"page path, e.g. /Home"`
	IncludeContent bool   `json:"includeContent,omitempty" jsonschema:"include the page content (optional)"`
	RecursionLevel string `json:"recursionLevel,omitempty" jsonschema:"recursion level for subpages, e.g. full (optional)"`
}

// CreateOrUpdatePageInput creates or updates a wiki page.
type CreateOrUpdatePageInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
	Path    string `json:"path" jsonschema:"page path, e.g. /Home"`
	Content string `json:"content" jsonschema:"the markdown content of the page"`
}

// CreateWikiInput creates a new wiki.
type CreateWikiInput struct {
	Project   string `json:"project" jsonschema:"project name or ID"`
	Name      string `json:"name" jsonschema:"the name of the wiki"`
	ProjectID string `json:"projectId" jsonschema:"the project ID the wiki belongs to"`
	Type      string `json:"type" jsonschema:"the wiki type: projectWiki or codeWiki"`
}

// DeletePageInput identifies a wiki page to delete.
type DeletePageInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
	Path    string `json:"path" jsonschema:"page path, e.g. /Home"`
}

// --- Tool handlers ---

func (s *service) listWikis(ctx context.Context, _ *mcp.CallToolRequest, in ListWikisInput) (*mcp.CallToolResult, server.ListResult[Wiki], error) {
	out, err := s.ListWikis(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) getWiki(ctx context.Context, _ *mcp.CallToolRequest, in GetWikiInput) (*mcp.CallToolResult, *Wiki, error) {
	out, err := s.GetWiki(ctx, in.Project, in.WikiID)
	return nil, out, err
}

func (s *service) getPage(ctx context.Context, _ *mcp.CallToolRequest, in GetPageInput) (*mcp.CallToolResult, *WikiPage, error) {
	out, err := s.GetPage(ctx, in.Project, in.WikiID, in.Path, in.IncludeContent, in.RecursionLevel)
	return nil, out, err
}

func (s *service) createOrUpdatePage(ctx context.Context, _ *mcp.CallToolRequest, in CreateOrUpdatePageInput) (*mcp.CallToolResult, *WikiPage, error) {
	out, err := s.CreateOrUpdatePage(ctx, in.Project, in.WikiID, in.Path, in.Content)
	return nil, out, err
}

func (s *service) createWiki(ctx context.Context, _ *mcp.CallToolRequest, in CreateWikiInput) (*mcp.CallToolResult, *Wiki, error) {
	out, err := s.CreateWiki(ctx, in.Project, in.Name, in.ProjectID, in.Type)
	return nil, out, err
}

func (s *service) deletePage(ctx context.Context, _ *mcp.CallToolRequest, in DeletePageInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.DeletePage(ctx, in.Project, in.WikiID, in.Path); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

// CreateAttachmentInput uploads a wiki attachment.
type CreateAttachmentInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
	Name    string `json:"name" jsonschema:"attachment file name, e.g. diagram.png"`
	Content string `json:"content" jsonschema:"file content; binary data must be base64-encoded"`
}

// MovePageInput moves or renames a wiki page.
type MovePageInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
	Path    string `json:"path" jsonschema:"current page path"`
	NewPath string `json:"newPath" jsonschema:"new page path"`
}

// MovePageOutput wraps the page-move result.
type MovePageOutput struct {
	Page map[string]any `json:"page" jsonschema:"the moved page"`
}

func (s *service) createAttachment(ctx context.Context, _ *mcp.CallToolRequest, in CreateAttachmentInput) (*mcp.CallToolResult, *Attachment, error) {
	out, err := s.CreateAttachment(ctx, in.Project, in.WikiID, in.Name, in.Content)
	return nil, out, err
}

func (s *service) movePage(ctx context.Context, _ *mcp.CallToolRequest, in MovePageInput) (*mcp.CallToolResult, *MovePageOutput, error) {
	out, err := s.MovePage(ctx, in.Project, in.WikiID, in.Path, in.NewPath)
	if err != nil {
		return nil, nil, err
	}
	return nil, &MovePageOutput{Page: out}, nil
}

// GetPageByIDInput identifies a page by numeric ID.
type GetPageByIDInput struct {
	Project        string `json:"project" jsonschema:"project name or ID"`
	WikiID         string `json:"wikiId" jsonschema:"wiki ID or name"`
	PageID         int    `json:"pageId" jsonschema:"numeric page ID"`
	IncludeContent bool   `json:"includeContent,omitempty" jsonschema:"include the page content (optional)"`
}

// GetPageStatsInput requests page view stats.
type GetPageStatsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
	PageID  int    `json:"pageId" jsonschema:"numeric page ID"`
	Days    int    `json:"days,omitempty" jsonschema:"number of days of view stats to return (optional)"`
}

// UpdateWikiInput renames a wiki.
type UpdateWikiInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
	Name    string `json:"name" jsonschema:"new wiki name"`
}

// DeleteWikiInput identifies a wiki to delete.
type DeleteWikiInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	WikiID  string `json:"wikiId" jsonschema:"wiki ID or name"`
}

func (s *service) getPageByID(ctx context.Context, _ *mcp.CallToolRequest, in GetPageByIDInput) (*mcp.CallToolResult, *WikiPage, error) {
	out, err := s.GetPageByID(ctx, in.Project, in.WikiID, in.PageID, in.IncludeContent)
	return nil, out, err
}

func (s *service) getPageStats(ctx context.Context, _ *mcp.CallToolRequest, in GetPageStatsInput) (*mcp.CallToolResult, *PageStats, error) {
	out, err := s.GetPageStats(ctx, in.Project, in.WikiID, in.PageID, in.Days)
	return nil, out, err
}

func (s *service) updateWiki(ctx context.Context, _ *mcp.CallToolRequest, in UpdateWikiInput) (*mcp.CallToolResult, *Wiki, error) {
	out, err := s.UpdateWiki(ctx, in.Project, in.WikiID, in.Name)
	return nil, out, err
}

func (s *service) deleteWiki(ctx context.Context, _ *mcp.CallToolRequest, in DeleteWikiInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.DeleteWiki(ctx, in.Project, in.WikiID); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

// AppendToPageInput appends content to a wiki page.
type AppendToPageInput struct {
	Project   string `json:"project" jsonschema:"project name or ID"`
	WikiID    string `json:"wikiId" jsonschema:"wiki ID or name"`
	Path      string `json:"path" jsonschema:"page path, e.g. /Changelog"`
	Content   string `json:"content" jsonschema:"Markdown content to append"`
	Separator string `json:"separator,omitempty" jsonschema:"separator between existing and new content (default two newlines)"`
}

// ListPagesInput lists a wiki page tree from a root path.
type ListPagesInput struct {
	Project  string `json:"project" jsonschema:"project name or ID"`
	WikiID   string `json:"wikiId" jsonschema:"wiki ID or name"`
	RootPath string `json:"rootPath,omitempty" jsonschema:"root path to list from (default /)"`
}

func (s *service) appendToPage(ctx context.Context, _ *mcp.CallToolRequest, in AppendToPageInput) (*mcp.CallToolResult, *WikiPage, error) {
	out, err := s.AppendToPage(ctx, in.Project, in.WikiID, in.Path, in.Content, in.Separator)
	return nil, out, err
}

func (s *service) listPages(ctx context.Context, _ *mcp.CallToolRequest, in ListPagesInput) (*mcp.CallToolResult, *WikiPage, error) {
	out, err := s.ListPages(ctx, in.Project, in.WikiID, in.RootPath)
	return nil, out, err
}
