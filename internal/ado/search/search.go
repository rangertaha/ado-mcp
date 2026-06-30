// SPDX-License-Identifier: MIT

// Package search exposes the Azure DevOps Search service: full-text search
// across code, work items, and wiki pages. All operations are POST requests
// against the almsearch.dev.azure.com host and are read-only.
package search

import (
	"context"

	"github.com/rangertaha/ado-mcp/internal/ado"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "search"

// service wraps the Azure DevOps clients for Search operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// searchRequest is the request body shared by all search endpoints.
type searchRequest struct {
	SearchText string              `json:"searchText"`
	Top        int                 `json:"$top"`
	Skip       int                 `json:"$skip"`
	Filters    map[string][]string `json:"filters,omitempty"`
}

// newRequest builds a search request body, defaulting top to 50 and scoping
// to the given project when one is supplied.
func newRequest(searchText, project string, top int) searchRequest {
	if top <= 0 {
		top = 50
	}
	req := searchRequest{SearchText: searchText, Top: top, Skip: 0}
	if project != "" {
		req.Filters = map[string][]string{"Project": {project}}
	}
	return req
}

// CodeSearchResponse is the envelope returned by the code search endpoint.
type CodeSearchResponse struct {
	Count   int `json:"count"`
	Results any `json:"results"`
}

// WorkItemSearchResponse is the envelope returned by the work item search endpoint.
type WorkItemSearchResponse struct {
	Count   int `json:"count"`
	Results any `json:"results"`
}

// WikiSearchResponse is the envelope returned by the wiki search endpoint.
type WikiSearchResponse struct {
	Count   int `json:"count"`
	Results any `json:"results"`
}

// --- Search operations ---

// SearchCode performs a full-text search across source code. project is
// optional and scopes the search to a single project; top optionally limits
// the number of results (default 50).
func (s *service) SearchCode(ctx context.Context, searchText, project string, top int) (*CodeSearchResponse, error) {
	body := newRequest(searchText, project, top)
	var out CodeSearchResponse
	if err := s.c.Search.PostJSONVersion(ctx, "/_apis/search/codesearchresults", body, &out, "7.1-preview.1"); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchWorkItems performs a full-text search across work items. project is
// optional and scopes the search to a single project; top optionally limits
// the number of results (default 50).
func (s *service) SearchWorkItems(ctx context.Context, searchText, project string, top int) (*WorkItemSearchResponse, error) {
	body := newRequest(searchText, project, top)
	var out WorkItemSearchResponse
	if err := s.c.Search.PostJSONVersion(ctx, "/_apis/search/workitemsearchresults", body, &out, "7.1-preview.1"); err != nil {
		return nil, err
	}
	return &out, nil
}

// SearchWiki performs a full-text search across wiki pages. project is
// optional and scopes the search to a single project; top optionally limits
// the number of results (default 50).
func (s *service) SearchWiki(ctx context.Context, searchText, project string, top int) (*WikiSearchResponse, error) {
	body := newRequest(searchText, project, top)
	var out WikiSearchResponse
	if err := s.c.Search.PostJSONVersion(ctx, "/_apis/search/wikisearchresults", body, &out, "7.1-preview.1"); err != nil {
		return nil, err
	}
	return &out, nil
}
