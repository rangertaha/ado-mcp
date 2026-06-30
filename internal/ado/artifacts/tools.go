// SPDX-License-Identifier: MIT

package artifacts

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Artifacts (packaging feeds) toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "artifacts_list_feeds", Title: "List feeds",
		Description: "List the Azure Artifacts packaging feeds available to the caller."}, svc.listFeeds)
	server.Register(s, server.ToolDef{Name: "artifacts_get_feed", Title: "Get feed",
		Description: "Get a single packaging feed by its ID or name."}, svc.getFeed)
	server.Register(s, server.ToolDef{Name: "artifacts_list_packages", Title: "List packages",
		Description: "List the packages contained in a feed."}, svc.listPackages)
	server.Register(s, server.ToolDef{Name: "artifacts_list_package_versions", Title: "List package versions",
		Description: "List the published versions of a package in a feed."}, svc.listPackageVersions)
	server.Register(s, server.ToolDef{Name: "artifacts_get_package", Title: "Get package",
		Description: "Get a single package within a feed by its ID."}, svc.getPackage)
	server.Register(s, server.ToolDef{Name: "artifacts_list_feed_views", Title: "List feed views",
		Description: "List the views defined on a packaging feed."}, svc.listFeedViews)
}

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// ListFeedsInput has no parameters.
type ListFeedsInput = EmptyInput

func (s *service) listFeeds(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[Feed], error) {
	out, err := s.ListFeeds(ctx)
	return nil, server.List(out), err
}

// GetFeedInput identifies a feed to fetch.
type GetFeedInput struct {
	FeedID string `json:"feedId" jsonschema:"feed ID or name"`
}

func (s *service) getFeed(ctx context.Context, _ *mcp.CallToolRequest, in GetFeedInput) (*mcp.CallToolResult, *Feed, error) {
	out, err := s.GetFeed(ctx, in.FeedID)
	return nil, out, err
}

// ListPackagesInput selects a feed and optionally limits the result count.
type ListPackagesInput struct {
	FeedID string `json:"feedId" jsonschema:"feed ID or name"`
	Top    int    `json:"top,omitempty" jsonschema:"maximum number of packages to return (optional)"`
}

func (s *service) listPackages(ctx context.Context, _ *mcp.CallToolRequest, in ListPackagesInput) (*mcp.CallToolResult, server.ListResult[Package], error) {
	out, err := s.ListPackages(ctx, in.FeedID, in.Top)
	return nil, server.List(out), err
}

// ListPackageVersionsInput identifies a package within a feed.
type ListPackageVersionsInput struct {
	FeedID    string `json:"feedId" jsonschema:"feed ID or name"`
	PackageID string `json:"packageId" jsonschema:"package ID"`
}

func (s *service) listPackageVersions(ctx context.Context, _ *mcp.CallToolRequest, in ListPackageVersionsInput) (*mcp.CallToolResult, server.ListResult[PackageVersion], error) {
	out, err := s.ListPackageVersions(ctx, in.FeedID, in.PackageID)
	return nil, server.List(out), err
}

// GetPackageInput identifies a package within a feed to fetch.
type GetPackageInput struct {
	FeedID    string `json:"feedId" jsonschema:"feed ID or name"`
	PackageID string `json:"packageId" jsonschema:"package ID"`
}

func (s *service) getPackage(ctx context.Context, _ *mcp.CallToolRequest, in GetPackageInput) (*mcp.CallToolResult, *Package, error) {
	out, err := s.GetPackage(ctx, in.FeedID, in.PackageID)
	return nil, out, err
}

// ListFeedViewsInput identifies a feed whose views to list.
type ListFeedViewsInput struct {
	FeedID string `json:"feedId" jsonschema:"feed ID or name"`
}

func (s *service) listFeedViews(ctx context.Context, _ *mcp.CallToolRequest, in ListFeedViewsInput) (*mcp.CallToolResult, server.ListResult[FeedView], error) {
	out, err := s.ListFeedViews(ctx, in.FeedID)
	return nil, server.List(out), err
}
