// SPDX-License-Identifier: MIT

// Package artifacts wraps the Azure DevOps Artifacts (packaging feeds) APIs.
//
// It exposes read-only operations for listing feeds, inspecting a single feed,
// and browsing the packages and package versions hosted in a feed. All requests
// target the feeds host (feeds.dev.azure.com) via the Feeds client.
package artifacts

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset identifier for the artifacts area.
const Name = "artifacts"

// service holds the Azure DevOps clients used by the artifacts operations.
type service struct{ c *ado.Clients }

// Feed describes an Azure Artifacts packaging feed.
type Feed struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description,omitempty"`
	URL                string `json:"url,omitempty"`
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
}

// Package describes a package stored within a feed.
type Package struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProtocolType string `json:"protocolType,omitempty"`
	URL          string `json:"url,omitempty"`
}

// PackageVersion describes a single published version of a package.
type PackageVersion struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	PublishDate string `json:"publishDate,omitempty"`
	IsLatest    bool   `json:"isLatest,omitempty"`
}

// FeedView describes a view defined on a packaging feed.
type FeedView struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type,omitempty"`
	URL        string `json:"url,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

// ListFeeds returns all packaging feeds the caller can access.
func (s *service) ListFeeds(ctx context.Context) ([]Feed, error) {
	var out client.List[Feed]
	if err := s.c.Feeds.GetJSON(ctx, "/_apis/packaging/feeds", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetFeed returns a single feed by its ID or name.
func (s *service) GetFeed(ctx context.Context, feedID string) (*Feed, error) {
	var out Feed
	path := fmt.Sprintf("/_apis/packaging/feeds/%s", url.PathEscape(feedID))
	if err := s.c.Feeds.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPackages returns the packages contained in a feed. When top > 0 the
// number of returned packages is limited accordingly.
func (s *service) ListPackages(ctx context.Context, feedID string, top int) ([]Package, error) {
	q := url.Values{}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[Package]
	path := fmt.Sprintf("/_apis/packaging/Feeds/%s/packages", url.PathEscape(feedID))
	if err := s.c.Feeds.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListPackageVersions returns the published versions of a package in a feed.
func (s *service) ListPackageVersions(ctx context.Context, feedID, packageID string) ([]PackageVersion, error) {
	var out client.List[PackageVersion]
	path := fmt.Sprintf("/_apis/packaging/Feeds/%s/Packages/%s/versions",
		url.PathEscape(feedID), url.PathEscape(packageID))
	if err := s.c.Feeds.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetPackage returns a single package within a feed by its ID.
func (s *service) GetPackage(ctx context.Context, feedID, packageID string) (*Package, error) {
	var out Package
	path := fmt.Sprintf("/_apis/packaging/Feeds/%s/Packages/%s",
		url.PathEscape(feedID), url.PathEscape(packageID))
	if err := s.c.Feeds.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListFeedViews returns the views defined on a packaging feed.
func (s *service) ListFeedViews(ctx context.Context, feedID string) ([]FeedView, error) {
	var out client.List[FeedView]
	path := fmt.Sprintf("/_apis/packaging/Feeds/%s/views", url.PathEscape(feedID))
	if err := s.c.Feeds.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
