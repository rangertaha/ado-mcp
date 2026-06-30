// SPDX-License-Identifier: MIT

// Package macros provides composite tools: single MCP tools that orchestrate
// several Azure DevOps REST calls server-side. They exist for common multi-step
// operations that are awkward or error-prone to drive one tool at a time
// (e.g. completing a pull request requires first fetching its last merge
// commit).
package macros

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "macros"

// service wraps the Azure DevOps clients for composite operations.
type service struct {
	c *ado.Clients
}

// pullRequest is the subset of PR fields the macros need.
type pullRequest struct {
	PullRequestID         int    `json:"pullRequestId"`
	Title                 string `json:"title,omitempty"`
	Status                string `json:"status,omitempty"`
	SourceRefName         string `json:"sourceRefName,omitempty"`
	TargetRefName         string `json:"targetRefName,omitempty"`
	URL                   string `json:"url,omitempty"`
	LastMergeSourceCommit *struct {
		CommitID string `json:"commitId"`
	} `json:"lastMergeSourceCommit,omitempty"`
}

// workItem is the subset of work item fields the macros return.
type workItem struct {
	ID     int            `json:"id"`
	URL    string         `json:"url,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
}

// patchOp is a JSON Patch operation for work item creation.
type patchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// CompletePullRequest fetches a PR to obtain its last merge source commit, then
// completes (merges) it with the requested merge options.
func (s *service) CompletePullRequest(ctx context.Context, project, repo string, id int, deleteSourceBranch, squash bool, mergeStrategy string) (*pullRequest, error) {
	base := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullrequests/%d", url.PathEscape(project), url.PathEscape(repo), id)

	var pr pullRequest
	if err := s.c.Org.GetJSON(ctx, base, nil, &pr); err != nil {
		return nil, fmt.Errorf("loading pull request: %w", err)
	}
	if pr.LastMergeSourceCommit == nil || pr.LastMergeSourceCommit.CommitID == "" {
		return nil, fmt.Errorf("pull request %d has no last merge source commit; it may not be mergeable", id)
	}

	completion := map[string]any{"deleteSourceBranch": deleteSourceBranch}
	if squash {
		completion["squashMerge"] = true
	}
	if mergeStrategy != "" {
		completion["mergeStrategy"] = mergeStrategy
	}
	body := map[string]any{
		"status":                "completed",
		"lastMergeSourceCommit": map[string]any{"commitId": pr.LastMergeSourceCommit.CommitID},
		"completionOptions":     completion,
	}

	var out pullRequest
	if err := s.c.Org.PatchJSON(ctx, base, nil, body, &out, ""); err != nil {
		return nil, fmt.Errorf("completing pull request: %w", err)
	}
	return &out, nil
}

// CreateBug creates a Bug work item from common fields and, optionally, adds a
// comment in the same operation.
func (s *service) CreateBug(ctx context.Context, project, title, repro, severity, assignedTo, areaPath, iterationPath, comment string) (*workItem, error) {
	ops := []patchOp{{Op: "add", Path: "/fields/System.Title", Value: title}}
	add := func(field, value string) {
		if value != "" {
			ops = append(ops, patchOp{Op: "add", Path: "/fields/" + field, Value: value})
		}
	}
	add("Microsoft.VSTS.TCM.ReproSteps", repro)
	add("Microsoft.VSTS.Common.Severity", severity)
	add("System.AssignedTo", assignedTo)
	add("System.AreaPath", areaPath)
	add("System.IterationPath", iterationPath)

	createPath := fmt.Sprintf("/%s/_apis/wit/workitems/$Bug", url.PathEscape(project))
	var wi workItem
	if err := s.c.Org.PostJSONPatch(ctx, createPath, ops, &wi); err != nil {
		return nil, fmt.Errorf("creating bug: %w", err)
	}

	if comment != "" {
		commentPath := fmt.Sprintf("/%s/_apis/wit/workItems/%d/comments", url.PathEscape(project), wi.ID)
		if err := s.c.Org.PostJSONVersion(ctx, commentPath, map[string]string{"text": comment}, nil, "7.1-preview.4"); err != nil {
			return &wi, fmt.Errorf("bug %d created but adding comment failed: %w", wi.ID, err)
		}
	}
	return &wi, nil
}

// PublishWikiPage creates or updates a wiki page and returns the resulting page.
// It reads the page first to learn whether it exists and to capture its ETag,
// then supplies that ETag as If-Match when updating — Azure DevOps requires it
// to edit an existing page. New pages are created without one.
func (s *service) PublishWikiPage(ctx context.Context, project, wiki, path, content string) (created bool, page map[string]any, err error) {
	base := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/pages", url.PathEscape(project), url.PathEscape(wiki))

	q := url.Values{}
	q.Set("path", path)
	q.Set("includeContent", "true")
	existed := true
	etag := ""
	var existing map[string]any
	if resp, e := s.c.Org.Do(ctx, client.Request{Method: "GET", Path: base, Query: q, Out: &existing}); e != nil {
		existed = false // most likely a 404: the page does not exist yet
	} else if resp != nil {
		etag = resp.Header.Get("ETag")
	}

	put := url.Values{}
	put.Set("path", path)
	header := http.Header{}
	if existed && etag != "" {
		header.Set("If-Match", etag)
	}
	var result map[string]any
	if _, err = s.c.Org.Do(ctx, client.Request{
		Method: "PUT", Path: base, Query: put,
		Body: map[string]string{"content": content}, Header: header, Out: &result,
	}); err != nil {
		return false, nil, fmt.Errorf("publishing wiki page: %w", err)
	}
	return !existed, result, nil
}

// WikiImage is an image to upload alongside a wiki page. Content is the raw file
// bytes, base64-encoded.
type WikiImage struct {
	Name    string
	Content string
}

// uploadedImage records an uploaded attachment and the path to reference it.
type uploadedImage struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// PublishWikiPageWithImages uploads images as wiki attachments, substitutes any
// "{{att:NAME}}" placeholders in the page content with each image's attachment
// path, then creates or updates the page (ETag-aware). This makes publishing a
// page that embeds images a single operation.
func (s *service) PublishWikiPageWithImages(ctx context.Context, project, wiki, path, content string, images []WikiImage) (created bool, page map[string]any, atts []uploadedImage, err error) {
	for _, img := range images {
		q := url.Values{}
		q.Set("name", img.Name)
		ap := fmt.Sprintf("/%s/_apis/wiki/wikis/%s/attachments", url.PathEscape(project), url.PathEscape(wiki))
		var ref struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}
		if _, e := s.c.Org.Do(ctx, client.Request{
			Method: "PUT", Path: ap, Query: q,
			Body: strings.NewReader(img.Content), ContentType: "application/octet-stream", Out: &ref,
		}); e != nil {
			return false, nil, nil, fmt.Errorf("uploading attachment %q: %w", img.Name, e)
		}
		refPath := ref.Path
		if refPath == "" {
			refPath = "/.attachments/" + img.Name
		}
		// Replace placeholders so the page can reference the image by name.
		content = strings.ReplaceAll(content, "{{att:"+img.Name+"}}", refPath)
		atts = append(atts, uploadedImage{Name: img.Name, Path: refPath})
	}

	created, page, err = s.PublishWikiPage(ctx, project, wiki, path, content)
	if err != nil {
		return false, nil, atts, err
	}
	return created, page, atts, nil
}
