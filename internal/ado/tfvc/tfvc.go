// SPDX-License-Identifier: MIT

// Package tfvc exposes the Azure DevOps Team Foundation Version Control (TFVC)
// service: listing and fetching changesets, listing branches, and reading item
// content.
package tfvc

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "tfvc"

// service wraps the Azure DevOps clients for TFVC operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Changeset is a TFVC changeset.
type Changeset struct {
	ChangesetID int    `json:"changesetId"`
	Comment     string `json:"comment,omitempty"`
	Author      any    `json:"author,omitempty"`
	CreatedDate string `json:"createdDate,omitempty"`
	URL         string `json:"url,omitempty"`
}

// TfvcBranch is a TFVC branch.
type TfvcBranch struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	Owner       any    `json:"owner,omitempty"`
	CreatedDate string `json:"createdDate,omitempty"`
}

// TfvcItem is a TFVC item (file or folder), optionally including its content.
type TfvcItem struct {
	Path     string `json:"path"`
	Content  string `json:"content,omitempty"`
	Version  int    `json:"version,omitempty"`
	IsFolder bool   `json:"isFolder,omitempty"`
	URL      string `json:"url,omitempty"`
}

// --- TFVC operations ---

// ListChangesets returns TFVC changesets, optionally filtered by item path.
func (s *service) ListChangesets(ctx context.Context, itemPath string, top int) ([]Changeset, error) {
	q := url.Values{}
	if itemPath != "" {
		q.Set("searchCriteria.itemPath", itemPath)
	}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[Changeset]
	if err := s.c.Org.GetJSON(ctx, "/_apis/tfvc/changesets", q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetChangeset returns a single TFVC changeset by ID.
func (s *service) GetChangeset(ctx context.Context, id int) (*Changeset, error) {
	var cs Changeset
	path := fmt.Sprintf("/_apis/tfvc/changesets/%s", url.PathEscape(strconv.Itoa(id)))
	if err := s.c.Org.GetJSON(ctx, path, nil, &cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// ListBranches returns the TFVC branches, including child branches.
func (s *service) ListBranches(ctx context.Context) ([]TfvcBranch, error) {
	q := url.Values{}
	q.Set("includeChildren", "true")
	var out client.List[TfvcBranch]
	if err := s.c.Org.GetJSON(ctx, "/_apis/tfvc/branches", q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetItem returns a TFVC item at the given path, including its content.
func (s *service) GetItem(ctx context.Context, itemPath string) (*TfvcItem, error) {
	q := url.Values{}
	q.Set("path", itemPath)
	q.Set("includeContent", "true")
	var item TfvcItem
	if err := s.c.Org.GetJSON(ctx, "/_apis/tfvc/items", q, &item); err != nil {
		return nil, err
	}
	return &item, nil
}
