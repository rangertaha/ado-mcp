// SPDX-License-Identifier: MIT

// Package git exposes the Azure DevOps Git service: repositories, refs/branches,
// commits, file contents, pull requests, PR threads/comments and reviewer votes.
package git

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "git"

// service wraps the Azure DevOps clients for Git operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Repository is a Git repository.
type Repository struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	URL           string `json:"url,omitempty"`
	WebURL        string `json:"webUrl,omitempty"`
	Size          int64  `json:"size,omitempty"`
	RemoteURL     string `json:"remoteUrl,omitempty"`
}

// Ref is a Git reference (branch or tag).
type Ref struct {
	Name     string `json:"name"`
	ObjectID string `json:"objectId,omitempty"`
	Creator  any    `json:"creator,omitempty"`
}

// Commit is a Git commit.
type Commit struct {
	CommitID string `json:"commitId"`
	Comment  string `json:"comment,omitempty"`
	Author   any    `json:"author,omitempty"`
	URL      string `json:"url,omitempty"`
}

// Item is a file or folder in a repository.
type Item struct {
	Path     string `json:"path"`
	IsFolder bool   `json:"isFolder,omitempty"`
	ObjectID string `json:"objectId,omitempty"`
	CommitID string `json:"commitId,omitempty"`
	Content  string `json:"content,omitempty"`
	URL      string `json:"url,omitempty"`
}

// PullRequest is a pull request.
type PullRequest struct {
	PullRequestID int    `json:"pullRequestId"`
	Title         string `json:"title,omitempty"`
	Description   string `json:"description,omitempty"`
	Status        string `json:"status,omitempty"`
	SourceRefName string `json:"sourceRefName,omitempty"`
	TargetRefName string `json:"targetRefName,omitempty"`
	CreatedBy     any    `json:"createdBy,omitempty"`
	URL           string `json:"url,omitempty"`
	IsDraft       bool   `json:"isDraft,omitempty"`
	MergeStatus   string `json:"mergeStatus,omitempty"`
}

// Thread is a pull request comment thread.
type Thread struct {
	ID       int         `json:"id"`
	Status   string      `json:"status,omitempty"`
	Comments []PRComment `json:"comments,omitempty"`
}

// PRComment is a single comment within a thread.
type PRComment struct {
	ID            int    `json:"id,omitempty"`
	ParentID      int    `json:"parentCommentId,omitempty"`
	Content       string `json:"content,omitempty"`
	Author        any    `json:"author,omitempty"`
	PublishedDate string `json:"publishedDate,omitempty"`
}

// CommitStatus is a status posted against a commit (e.g. a CI build result).
type CommitStatus struct {
	State       string `json:"state,omitempty"`
	Description string `json:"description,omitempty"`
	TargetURL   string `json:"targetUrl,omitempty"`
	Context     any    `json:"context,omitempty"`
	CreatedDate string `json:"createdDate,omitempty"`
}

// UpdateRefResult is the result of creating or updating a ref.
type UpdateRefResult struct {
	Name        string `json:"name,omitempty"`
	Success     bool   `json:"success,omitempty"`
	NewObjectID string `json:"newObjectId,omitempty"`
}

// --- Repository operations ---

// CreateRepository creates a new Git repository in a project.
func (s *service) CreateRepository(ctx context.Context, project, name string) (*Repository, error) {
	var r Repository
	body := map[string]any{"name": name}
	path := fmt.Sprintf("/%s/_apis/git/repositories", url.PathEscape(project))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// DeleteRepository deletes a Git repository by its ID.
func (s *service) DeleteRepository(ctx context.Context, repoID string) error {
	path := fmt.Sprintf("/_apis/git/repositories/%s", url.PathEscape(repoID))
	return s.c.Org.Delete(ctx, path, nil, nil)
}

// ListCommitStatuses returns the statuses posted against a commit.
func (s *service) ListCommitStatuses(ctx context.Context, project, repo, commitID string) ([]CommitStatus, error) {
	var out client.List[CommitStatus]
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/commits/%s/statuses",
		url.PathEscape(project), url.PathEscape(repo), url.PathEscape(commitID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// CreateCommitStatus posts a new status against a commit.
func (s *service) CreateCommitStatus(ctx context.Context, project, repo, commitID string, body map[string]any) (*CommitStatus, error) {
	var st CommitStatus
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/commits/%s/statuses",
		url.PathEscape(project), url.PathEscape(repo), url.PathEscape(commitID))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

// CreateBranch creates a new branch pointing at the given commit SHA.
func (s *service) CreateBranch(ctx context.Context, project, repo, branch, sha string) (*UpdateRefResult, error) {
	body := []map[string]any{{
		"name":        "refs/heads/" + branch,
		"oldObjectId": "0000000000000000000000000000000000000000",
		"newObjectId": sha,
	}}
	var out client.List[UpdateRefResult]
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/refs", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	if len(out.Value) == 0 {
		return nil, fmt.Errorf("create branch returned no results")
	}
	return &out.Value[0], nil
}

// ListRepositories returns the repositories in a project.
func (s *service) ListRepositories(ctx context.Context, project string) ([]Repository, error) {
	var out client.List[Repository]
	path := fmt.Sprintf("/%s/_apis/git/repositories", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetRepository returns a single repository by name or ID.
func (s *service) GetRepository(ctx context.Context, project, repo string) (*Repository, error) {
	var r Repository
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.GetJSON(ctx, path, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// ListRefs returns refs (branches/tags) in a repository, optionally filtered
// (e.g. filter="heads/" for branches, "tags/" for tags).
func (s *service) ListRefs(ctx context.Context, project, repo, filter string) ([]Ref, error) {
	q := url.Values{}
	if filter != "" {
		q.Set("filter", filter)
	}
	var out client.List[Ref]
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/refs", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListCommits returns commits in a repository, optionally restricted to a branch.
func (s *service) ListCommits(ctx context.Context, project, repo, branch string, top int) ([]Commit, error) {
	q := url.Values{}
	if branch != "" {
		q.Set("searchCriteria.itemVersion.version", branch)
	}
	if top > 0 {
		q.Set("searchCriteria.$top", strconv.Itoa(top))
	}
	var out client.List[Commit]
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/commits", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetFile returns the text content of a file at a path, optionally at a branch.
func (s *service) GetFile(ctx context.Context, project, repo, path, branch string) (*Item, error) {
	q := url.Values{}
	q.Set("path", path)
	q.Set("includeContent", "true")
	if branch != "" {
		q.Set("versionDescriptor.version", branch)
		q.Set("versionDescriptor.versionType", "branch")
	}
	var item Item
	apiPath := fmt.Sprintf("/%s/_apis/git/repositories/%s/items", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.GetJSON(ctx, apiPath, q, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// --- Pull request operations ---

// ListPullRequests returns pull requests in a repository filtered by status
// ("active", "completed", "abandoned", or "all").
func (s *service) ListPullRequests(ctx context.Context, project, repo, status string, top int) ([]PullRequest, error) {
	q := url.Values{}
	if status != "" {
		q.Set("searchCriteria.status", status)
	}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[PullRequest]
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullrequests", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetPullRequest returns a single pull request by ID.
func (s *service) GetPullRequest(ctx context.Context, project, repo string, id int) (*PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullrequests/%d", url.PathEscape(project), url.PathEscape(repo), id)
	if err := s.c.Org.GetJSON(ctx, path, nil, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// CreatePullRequest opens a new pull request.
func (s *service) CreatePullRequest(ctx context.Context, project, repo string, body map[string]any) (*PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullrequests", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// UpdatePullRequest patches a pull request (e.g. status, title, description).
func (s *service) UpdatePullRequest(ctx context.Context, project, repo string, id int, body map[string]any) (*PullRequest, error) {
	var pr PullRequest
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullrequests/%d", url.PathEscape(project), url.PathEscape(repo), id)
	if err := s.c.Org.PatchJSON(ctx, path, nil, body, &pr, ""); err != nil {
		return nil, err
	}
	return &pr, nil
}

// ListThreads returns the comment threads on a pull request.
func (s *service) ListThreads(ctx context.Context, project, repo string, id int) ([]Thread, error) {
	var out client.List[Thread]
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullRequests/%d/threads", url.PathEscape(project), url.PathEscape(repo), id)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// AddComment posts a new comment thread to a pull request.
func (s *service) AddComment(ctx context.Context, project, repo string, id int, content string) (*Thread, error) {
	body := map[string]any{
		"comments": []map[string]any{{"parentCommentId": 0, "content": content, "commentType": 1}},
		"status":   1, // active
	}
	var t Thread
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullRequests/%d/threads", url.PathEscape(project), url.PathEscape(repo), id)
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// SetVote sets the current/identified reviewer's vote on a pull request.
// Vote values: 10 approved, 5 approved-with-suggestions, 0 none, -5 waiting, -10 rejected.
func (s *service) SetVote(ctx context.Context, project, repo string, id int, reviewerID string, vote int) error {
	body := map[string]any{"vote": vote}
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullRequests/%d/reviewers/%s",
		url.PathEscape(project), url.PathEscape(repo), id, url.PathEscape(reviewerID))
	return s.c.Org.PutJSON(ctx, path, nil, body, nil)
}

// Reviewer is a reviewer on a pull request.
type Reviewer struct {
	ID          string `json:"id,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	UniqueName  string `json:"uniqueName,omitempty"`
	Vote        int    `json:"vote"`
	IsRequired  bool   `json:"isRequired,omitempty"`
	IsFlagged   bool   `json:"isFlagged,omitempty"`
}

// ListReviewers returns the reviewers on a pull request.
func (s *service) ListReviewers(ctx context.Context, project, repo string, id int) ([]Reviewer, error) {
	var out client.List[Reviewer]
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullRequests/%d/reviewers",
		url.PathEscape(project), url.PathEscape(repo), id)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// AddReviewer adds (or updates) a reviewer on a pull request, optionally marking
// them required.
func (s *service) AddReviewer(ctx context.Context, project, repo string, id int, reviewerID string, required bool) (*Reviewer, error) {
	body := map[string]any{"isRequired": required, "id": reviewerID}
	var r Reviewer
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullRequests/%d/reviewers/%s",
		url.PathEscape(project), url.PathEscape(repo), id, url.PathEscape(reviewerID))
	if err := s.c.Org.PutJSON(ctx, path, nil, body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// RemoveReviewer removes a reviewer from a pull request.
func (s *service) RemoveReviewer(ctx context.Context, project, repo string, id int, reviewerID string) error {
	path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullRequests/%d/reviewers/%s",
		url.PathEscape(project), url.PathEscape(repo), id, url.PathEscape(reviewerID))
	return s.c.Org.Delete(ctx, path, nil, nil)
}

// Push is the result of a git push (a new commit pushed to a branch).
type Push struct {
	PushID  int    `json:"pushId"`
	Date    string `json:"date,omitempty"`
	URL     string `json:"url,omitempty"`
	Commits []struct {
		CommitID string `json:"commitId"`
		Comment  string `json:"comment,omitempty"`
	} `json:"commits,omitempty"`
}

// branchHeadObjectID resolves a branch name to its current tip commit SHA, used
// as the oldObjectId when pushing. Returns the zero SHA when the branch does not
// yet exist (so a push can create it).
func (s *service) branchHeadObjectID(ctx context.Context, project, repo, branch string) (string, error) {
	refs, err := s.ListRefs(ctx, project, repo, "heads/"+branch)
	if err != nil {
		return "", err
	}
	want := "refs/heads/" + branch
	for _, r := range refs {
		if r.Name == want {
			return r.ObjectID, nil
		}
	}
	return "0000000000000000000000000000000000000000", nil
}

// PushFile commits a single file change (add, edit, or delete) to a branch in
// one push. It resolves the branch tip automatically for the required
// oldObjectId. changeType must be "add", "edit", or "delete".
func (s *service) PushFile(ctx context.Context, project, repo, branch, path, content, message, changeType string) (*Push, error) {
	if changeType == "" {
		changeType = "add"
	}
	oldObjectID, err := s.branchHeadObjectID(ctx, project, repo, branch)
	if err != nil {
		return nil, fmt.Errorf("resolving branch %q: %w", branch, err)
	}

	change := map[string]any{
		"changeType": changeType,
		"item":       map[string]any{"path": path},
	}
	if changeType != "delete" {
		change["newContent"] = map[string]any{"content": content, "contentType": "rawtext"}
	}
	body := map[string]any{
		"refUpdates": []map[string]any{{"name": "refs/heads/" + branch, "oldObjectId": oldObjectID}},
		"commits":    []map[string]any{{"comment": message, "changes": []map[string]any{change}}},
	}

	var push Push
	apiPath := fmt.Sprintf("/%s/_apis/git/repositories/%s/pushes", url.PathEscape(project), url.PathEscape(repo))
	if err := s.c.Org.PostJSON(ctx, apiPath, nil, body, &push); err != nil {
		return nil, err
	}
	return &push, nil
}
