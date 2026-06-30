// SPDX-License-Identifier: MIT

package macros

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the composite "macro" toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{
		Name:        "macro_complete_pull_request",
		Title:       "Complete pull request",
		Description: "Complete (merge) a pull request. Fetches the PR's last merge commit automatically, then merges with the chosen options.",
		Write:       true,
		Idempotent:  true,
	}, svc.completePullRequest)

	server.Register(s, server.ToolDef{
		Name:        "macro_create_bug",
		Title:       "Create bug",
		Description: "Create a Bug work item from common fields (title, repro steps, severity, assignee, area/iteration), optionally adding a comment in the same call.",
		Write:       true,
	}, svc.createBug)

	server.Register(s, server.ToolDef{
		Name:        "macro_publish_wiki_page",
		Title:       "Publish wiki page",
		Description: "Create or update a wiki page at a path with the given Markdown content, reporting whether it was created or updated.",
		Write:       true,
		Idempotent:  true,
	}, svc.publishWikiPage)

	server.Register(s, server.ToolDef{
		Name:  "macro_publish_wiki_page_with_images",
		Title: "Publish wiki page with images",
		Description: "Upload images as wiki attachments and create/update a page that embeds them. " +
			"In the Markdown content, reference each image with the placeholder {{att:NAME}} " +
			"(NAME matching an image's name); it is replaced with the attachment path. " +
			"Image content must be base64-encoded.",
		Write:      true,
		Idempotent: true,
	}, svc.publishWikiPageWithImages)
}

// --- Tool input/output types ---

// CompletePullRequestInput merges a pull request.
type CompletePullRequestInput struct {
	Project            string `json:"project" jsonschema:"project name or ID"`
	Repo               string `json:"repo" jsonschema:"repository name or ID"`
	PullRequestID      int    `json:"pullRequestId" jsonschema:"pull request ID"`
	DeleteSourceBranch bool   `json:"deleteSourceBranch,omitempty" jsonschema:"delete the source branch after merge (optional)"`
	Squash             bool   `json:"squash,omitempty" jsonschema:"squash the commits into one (optional)"`
	MergeStrategy      string `json:"mergeStrategy,omitempty" jsonschema:"merge strategy: noFastForward, squash, rebase, or rebaseMerge (optional)"`
}

// CreateBugInput describes a new bug.
type CreateBugInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	Title         string `json:"title" jsonschema:"bug title"`
	ReproSteps    string `json:"reproSteps,omitempty" jsonschema:"steps to reproduce (HTML or text, optional)"`
	Severity      string `json:"severity,omitempty" jsonschema:"severity, e.g. '2 - High' (optional)"`
	AssignedTo    string `json:"assignedTo,omitempty" jsonschema:"assignee (display name or email, optional)"`
	AreaPath      string `json:"areaPath,omitempty" jsonschema:"area path (optional)"`
	IterationPath string `json:"iterationPath,omitempty" jsonschema:"iteration path (optional)"`
	Comment       string `json:"comment,omitempty" jsonschema:"a comment to add after creation (optional)"`
}

// PublishWikiPageInput creates or updates a wiki page.
type PublishWikiPageInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Wiki    string `json:"wiki" jsonschema:"wiki ID or name"`
	Path    string `json:"path" jsonschema:"page path, e.g. /Onboarding/Setup"`
	Content string `json:"content" jsonschema:"the Markdown page content"`
}

// PublishWikiPageOutput reports the outcome of publishing a wiki page.
type PublishWikiPageOutput struct {
	Created bool           `json:"created" jsonschema:"true if the page was newly created, false if updated"`
	Page    map[string]any `json:"page" jsonschema:"the resulting wiki page"`
}

// --- Handlers ---

func (s *service) completePullRequest(ctx context.Context, _ *mcp.CallToolRequest, in CompletePullRequestInput) (*mcp.CallToolResult, *pullRequest, error) {
	out, err := s.CompletePullRequest(ctx, in.Project, in.Repo, in.PullRequestID, in.DeleteSourceBranch, in.Squash, in.MergeStrategy)
	return nil, out, err
}

func (s *service) createBug(ctx context.Context, _ *mcp.CallToolRequest, in CreateBugInput) (*mcp.CallToolResult, *workItem, error) {
	out, err := s.CreateBug(ctx, in.Project, in.Title, in.ReproSteps, in.Severity, in.AssignedTo, in.AreaPath, in.IterationPath, in.Comment)
	return nil, out, err
}

func (s *service) publishWikiPage(ctx context.Context, _ *mcp.CallToolRequest, in PublishWikiPageInput) (*mcp.CallToolResult, *PublishWikiPageOutput, error) {
	created, page, err := s.PublishWikiPage(ctx, in.Project, in.Wiki, in.Path, in.Content)
	if err != nil {
		return nil, nil, err
	}
	return nil, &PublishWikiPageOutput{Created: created, Page: page}, nil
}

// WikiImageInput is one image to upload with a page.
type WikiImageInput struct {
	Name    string `json:"name" jsonschema:"image file name, referenced in content as {{att:NAME}}"`
	Content string `json:"content" jsonschema:"base64-encoded image bytes"`
}

// PublishWikiPageWithImagesInput publishes a page that embeds uploaded images.
type PublishWikiPageWithImagesInput struct {
	Project string           `json:"project" jsonschema:"project name or ID"`
	Wiki    string           `json:"wiki" jsonschema:"wiki ID or name"`
	Path    string           `json:"path" jsonschema:"page path, e.g. /Design/Overview"`
	Content string           `json:"content" jsonschema:"Markdown content; reference images with {{att:NAME}} placeholders"`
	Images  []WikiImageInput `json:"images" jsonschema:"images to upload and embed"`
}

// PublishWikiPageWithImagesOutput reports the page and the uploaded attachments.
type PublishWikiPageWithImagesOutput struct {
	Created     bool            `json:"created" jsonschema:"true if the page was newly created"`
	Page        map[string]any  `json:"page" jsonschema:"the resulting wiki page"`
	Attachments []uploadedImage `json:"attachments" jsonschema:"uploaded images and their attachment paths"`
}

func (s *service) publishWikiPageWithImages(ctx context.Context, _ *mcp.CallToolRequest, in PublishWikiPageWithImagesInput) (*mcp.CallToolResult, *PublishWikiPageWithImagesOutput, error) {
	imgs := make([]WikiImage, len(in.Images))
	for i, im := range in.Images {
		imgs[i] = WikiImage(im)
	}
	created, page, atts, err := s.PublishWikiPageWithImages(ctx, in.Project, in.Wiki, in.Path, in.Content, imgs)
	if err != nil {
		return nil, nil, err
	}
	return nil, &PublishWikiPageWithImagesOutput{Created: created, Page: page, Attachments: atts}, nil
}
