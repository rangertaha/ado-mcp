// SPDX-License-Identifier: MIT

// Package wiki exposes the Azure DevOps Wiki service: listing wikis,
// fetching a wiki, reading wiki pages, and creating or updating pages.
package wiki

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "wiki"

// service wraps the Azure DevOps clients for Wiki operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Wiki is an Azure DevOps wiki (project or code wiki).
type Wiki struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	ProjectID    string `json:"projectId,omitempty"`
	RepositoryID string `json:"repositoryId,omitempty"`
	MappedPath   string `json:"mappedPath,omitempty"`
	RemoteURL    string `json:"remoteUrl,omitempty"`
	URL          string `json:"url,omitempty"`
}

// WikiPage is a page within a wiki, optionally including its content and subpages.
type WikiPage struct {
	Path         string `json:"path"`
	Content      string `json:"content,omitempty"`
	Order        int    `json:"order,omitempty"`
	IsParentPage bool   `json:"isParentPage,omitempty"`
	GitItemPath  string `json:"gitItemPath,omitempty"`
	URL          string `json:"url,omitempty"`
	SubPages     any    `json:"subPages,omitempty"`
}

// --- Wiki operations ---

// ListWikis returns the wikis in a project.
func (s *service) ListWikis(ctx context.Context, project string) ([]Wiki, error) {
	var out client.List[Wiki]
	path := fmt.Sprintf("/%s/_apis/wiki/wikis", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetWiki returns a single wiki by ID or name.
func (s *service) GetWiki(ctx context.Context, project, wikiID string) (*Wiki, error) {
	var w Wiki
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s", url.PathEscape(project), url.PathEscape(wikiID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// GetPage returns a wiki page, optionally including its content and subpages.
func (s *service) GetPage(ctx context.Context, project, wikiID, pagePath string, includeContent bool, recursionLevel string) (*WikiPage, error) {
	q := url.Values{}
	if pagePath != "" {
		q.Set("path", pagePath)
	}
	if includeContent {
		q.Set("includeContent", "true")
	}
	if recursionLevel != "" {
		q.Set("recursionLevel", recursionLevel)
	}
	var p WikiPage
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pages", url.PathEscape(project), url.PathEscape(wikiID))
	if err := s.c.Org.GetJSON(ctx, path, q, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// getPageWithETag fetches a page (with content) and returns it along with its
// ETag (the page version). A nil page with empty etag and no error means the
// page does not exist yet.
func (s *service) getPageWithETag(ctx context.Context, project, wikiID, pagePath string) (page *WikiPage, etag string, err error) {
	q := url.Values{}
	q.Set("path", pagePath)
	q.Set("includeContent", "true")
	base := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pages", url.PathEscape(project), url.PathEscape(wikiID))

	var p WikiPage
	resp, err := s.c.Org.Do(ctx, client.Request{Method: "GET", Path: base, Query: q, Out: &p})
	if err != nil {
		var apiErr *client.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			return nil, "", nil // page does not exist yet
		}
		return nil, "", err
	}
	return &p, resp.Header.Get("ETag"), nil
}

// CreateOrUpdatePage creates a new wiki page, or updates an existing one at the
// given path. Azure DevOps requires an If-Match ETag to edit an existing page,
// so this reads the page first to obtain its current version and supplies the
// ETag on update; new pages are created without one. This makes editing a page
// work reliably rather than failing with a precondition error.
func (s *service) CreateOrUpdatePage(ctx context.Context, project, wikiID, pagePath, content string) (*WikiPage, error) {
	_, etag, err := s.getPageWithETag(ctx, project, wikiID, pagePath)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("path", pagePath)
	base := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pages", url.PathEscape(project), url.PathEscape(wikiID))

	header := http.Header{}
	if etag != "" {
		header.Set("If-Match", etag) // update an existing page
	}

	var p WikiPage
	if _, err := s.c.Org.Do(ctx, client.Request{
		Method: "PUT", Path: base, Query: q,
		Body: map[string]any{"content": content}, Header: header, Out: &p,
	}); err != nil {
		return nil, err
	}
	return &p, nil
}

// AppendToPage appends content to an existing wiki page (creating it if absent),
// optionally inserting a separator between the old and new content. This makes
// incremental maintenance — adding a changelog entry, a note, a new section —
// possible without resending the whole page.
func (s *service) AppendToPage(ctx context.Context, project, wikiID, pagePath, content, separator string) (*WikiPage, error) {
	existing, _, err := s.getPageWithETag(ctx, project, wikiID, pagePath)
	if err != nil {
		return nil, err
	}
	combined := content
	if existing != nil && existing.Content != "" {
		sep := separator
		if sep == "" {
			sep = "\n\n"
		}
		combined = existing.Content + sep + content
	}
	return s.CreateOrUpdatePage(ctx, project, wikiID, pagePath, combined)
}

// ListPages returns the page tree of a wiki starting at rootPath (default "/"),
// expanded to full depth. It is a convenience over GetPage(recursionLevel=full)
// for browsing and maintaining wiki structure.
func (s *service) ListPages(ctx context.Context, project, wikiID, rootPath string) (*WikiPage, error) {
	if rootPath == "" {
		rootPath = "/"
	}
	return s.GetPage(ctx, project, wikiID, rootPath, false, "full")
}

// CreateWiki creates a new project or code wiki in a project. wikiType must be
// "projectWiki" or "codeWiki".
func (s *service) CreateWiki(ctx context.Context, project, name, projectID, wikiType string) (*Wiki, error) {
	body := map[string]any{"name": name, "projectId": projectID, "type": wikiType}
	var w Wiki
	path := fmt.Sprintf("/%s/_apis/wiki/wikis", url.PathEscape(project))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

// DeletePage deletes a wiki page at the given path.
func (s *service) DeletePage(ctx context.Context, project, wikiID, pagePath string) error {
	q := url.Values{}
	q.Set("path", pagePath)
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pages", url.PathEscape(project), url.PathEscape(wikiID))
	return s.c.Org.Delete(ctx, path, q, nil)
}

// Attachment is a wiki attachment reference.
type Attachment struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path,omitempty"`
}

// CreateAttachment uploads an attachment to a wiki. content is the raw file
// bytes; binary content should be supplied base64-encoded as required by the
// Azure DevOps wiki attachments API.
func (s *service) CreateAttachment(ctx context.Context, project, wikiID, name, content string) (*Attachment, error) {
	q := url.Values{}
	q.Set("name", name)
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/attachments", url.PathEscape(project), url.PathEscape(wikiID))
	var a Attachment
	_, err := s.c.Org.Do(ctx, client.Request{
		Method:      "PUT",
		Path:        path,
		Query:       q,
		Body:        strings.NewReader(content),
		ContentType: "application/octet-stream",
		Out:         &a,
	})
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// MovePage moves or renames a wiki page from one path to another.
func (s *service) MovePage(ctx context.Context, project, wikiID, fromPath, toPath string) (map[string]any, error) {
	body := map[string]string{"path": fromPath, "newPath": toPath}
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pagemoves", url.PathEscape(project), url.PathEscape(wikiID))
	var out map[string]any
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteWiki deletes (unpublishes) a wiki by ID or name.
func (s *service) DeleteWiki(ctx context.Context, project, wikiID string) error {
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s", url.PathEscape(project), url.PathEscape(wikiID))
	return s.c.Org.Delete(ctx, path, nil, nil)
}

// UpdateWiki renames a wiki.
func (s *service) UpdateWiki(ctx context.Context, project, wikiID, name string) (*Wiki, error) {
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s", url.PathEscape(project), url.PathEscape(wikiID))
	var w Wiki
	if err := s.c.Org.PatchJSON(ctx, path, nil, map[string]string{"name": name}, &w, ""); err != nil {
		return nil, err
	}
	return &w, nil
}

// GetPageByID returns a wiki page by its numeric page ID, optionally including
// content.
func (s *service) GetPageByID(ctx context.Context, project, wikiID string, pageID int, includeContent bool) (*WikiPage, error) {
	q := url.Values{}
	if includeContent {
		q.Set("includeContent", "true")
	}
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pages/%d", url.PathEscape(project), url.PathEscape(wikiID), pageID)
	var p WikiPage
	if err := s.c.Org.GetJSON(ctx, path, q, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// PageStats holds page view statistics.
type PageStats struct {
	Count     int `json:"count"`
	ViewStats any `json:"viewStats,omitempty"`
}

// GetPageStats returns view statistics for a page over the last N days.
func (s *service) GetPageStats(ctx context.Context, project, wikiID string, pageID, days int) (*PageStats, error) {
	q := url.Values{}
	if days > 0 {
		q.Set("pageViewsForDays", strconv.Itoa(days))
	}
	path := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pages/%d/stats", url.PathEscape(project), url.PathEscape(wikiID), pageID)
	var st PageStats
	if err := s.c.Org.GetJSON(ctx, path, q, &st); err != nil {
		return nil, err
	}
	return &st, nil
}
