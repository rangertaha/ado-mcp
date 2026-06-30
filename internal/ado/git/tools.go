// SPDX-License-Identifier: MIT

package git

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Git toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "git_list_repositories", Title: "List repositories",
		Description: "List the Git repositories in a project."}, svc.listRepositories)
	server.Register(s, server.ToolDef{Name: "git_get_repository", Title: "Get repository",
		Description: "Get a Git repository by name or ID."}, svc.getRepository)
	server.Register(s, server.ToolDef{Name: "git_list_refs", Title: "List refs/branches",
		Description: "List refs in a repository. Use filter \"heads/\" for branches or \"tags/\" for tags."}, svc.listRefs)
	server.Register(s, server.ToolDef{Name: "git_list_commits", Title: "List commits",
		Description: "List commits in a repository, optionally on a specific branch."}, svc.listCommits)
	server.Register(s, server.ToolDef{Name: "git_get_file", Title: "Get file content",
		Description: "Get the text content of a file at a path, optionally at a specific branch."}, svc.getFile)
	server.Register(s, server.ToolDef{Name: "git_list_pull_requests", Title: "List pull requests",
		Description: "List pull requests in a repository, filtered by status (active, completed, abandoned, all)."}, svc.listPullRequests)
	server.Register(s, server.ToolDef{Name: "git_get_pull_request", Title: "Get pull request",
		Description: "Get a pull request by ID."}, svc.getPullRequest)
	server.Register(s, server.ToolDef{Name: "git_list_pr_threads", Title: "List PR comment threads",
		Description: "List the comment threads on a pull request."}, svc.listThreads)

	// --- Write tools ---
	server.Register(s, server.ToolDef{Name: "git_create_pull_request", Title: "Create pull request",
		Description: "Create a pull request from a source branch into a target branch.", Write: true}, svc.createPullRequest)
	server.Register(s, server.ToolDef{Name: "git_update_pull_request", Title: "Update pull request",
		Description: "Update a pull request: set status to completed or abandoned, or change title/description.", Write: true, Idempotent: true}, svc.updatePullRequest)
	server.Register(s, server.ToolDef{Name: "git_add_pr_comment", Title: "Add PR comment",
		Description: "Add a new comment thread to a pull request.", Write: true}, svc.addComment)
	server.Register(s, server.ToolDef{Name: "git_vote_pull_request", Title: "Vote on pull request",
		Description: "Set a reviewer's vote on a pull request (10 approve, 5 approve with suggestions, 0 reset, -5 waiting, -10 reject).", Write: true, Idempotent: true}, svc.vote)

	server.Register(s, server.ToolDef{Name: "git_list_pr_reviewers", Title: "List PR reviewers",
		Description: "List the reviewers on a pull request and their votes."}, svc.listReviewers)
	server.Register(s, server.ToolDef{Name: "git_add_pr_reviewer", Title: "Add PR reviewer",
		Description: "Add (or update) a reviewer on a pull request, optionally marking them required.", Write: true, Idempotent: true}, svc.addReviewer)
	server.Register(s, server.ToolDef{Name: "git_remove_pr_reviewer", Title: "Remove PR reviewer",
		Description: "Remove a reviewer from a pull request.", Write: true, Destructive: true, Idempotent: true}, svc.removeReviewer)

	server.Register(s, server.ToolDef{Name: "git_push_file", Title: "Commit a file",
		Description: "Add, edit, or delete a single file on a branch in one commit (push). changeType is add, edit, or delete; the branch tip is resolved automatically.", Write: true}, svc.pushFile)

	server.Register(s, server.ToolDef{Name: "git_list_commit_statuses", Title: "List commit statuses",
		Description: "List the statuses posted against a commit."}, svc.listCommitStatuses)
	server.Register(s, server.ToolDef{Name: "git_create_repository", Title: "Create repository",
		Description: "Create a new Git repository in a project.", Write: true}, svc.createRepository)
	server.Register(s, server.ToolDef{Name: "git_delete_repository", Title: "Delete repository",
		Description: "Delete a Git repository by its ID.", Write: true, Destructive: true}, svc.deleteRepository)
	server.Register(s, server.ToolDef{Name: "git_create_commit_status", Title: "Create commit status",
		Description: "Post a status against a commit (state: succeeded, failed, pending, error, notApplicable).", Write: true}, svc.createCommitStatus)
	server.Register(s, server.ToolDef{Name: "git_create_branch", Title: "Create branch",
		Description: "Create a new branch pointing at a commit SHA.", Write: true}, svc.createBranch)
}

// --- Tool input types ---

// ProjectInput identifies a project.
type ProjectInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

// RepoInput identifies a repository within a project.
type RepoInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Repo    string `json:"repo" jsonschema:"repository name or ID"`
}

// ListRefsInput filters refs.
type ListRefsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Repo    string `json:"repo" jsonschema:"repository name or ID"`
	Filter  string `json:"filter,omitempty" jsonschema:"ref filter, e.g. heads/ for branches or tags/ for tags"`
}

// ListCommitsInput filters commits.
type ListCommitsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Repo    string `json:"repo" jsonschema:"repository name or ID"`
	Branch  string `json:"branch,omitempty" jsonschema:"branch name to list commits from (optional)"`
	Top     int    `json:"top,omitempty" jsonschema:"maximum number of commits (optional)"`
}

// GetFileInput identifies a file to read.
type GetFileInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Repo    string `json:"repo" jsonschema:"repository name or ID"`
	Path    string `json:"path" jsonschema:"file path, e.g. /src/main.go"`
	Branch  string `json:"branch,omitempty" jsonschema:"branch name (optional, defaults to the default branch)"`
}

// ListPullRequestsInput filters pull requests.
type ListPullRequestsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Repo    string `json:"repo" jsonschema:"repository name or ID"`
	Status  string `json:"status,omitempty" jsonschema:"status filter: active, completed, abandoned, or all"`
	Top     int    `json:"top,omitempty" jsonschema:"maximum number of pull requests (optional)"`
}

// PullRequestInput identifies a pull request.
type PullRequestInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	Repo          string `json:"repo" jsonschema:"repository name or ID"`
	PullRequestID int    `json:"pullRequestId" jsonschema:"pull request ID"`
}

// CreatePullRequestInput describes a new pull request.
type CreatePullRequestInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	Repo         string `json:"repo" jsonschema:"repository name or ID"`
	SourceBranch string `json:"sourceBranch" jsonschema:"source branch, e.g. refs/heads/feature or feature"`
	TargetBranch string `json:"targetBranch" jsonschema:"target branch, e.g. refs/heads/main or main"`
	Title        string `json:"title" jsonschema:"pull request title"`
	Description  string `json:"description,omitempty" jsonschema:"pull request description (optional)"`
	IsDraft      bool   `json:"isDraft,omitempty" jsonschema:"create as a draft (optional)"`
}

// UpdatePullRequestInput patches a pull request.
type UpdatePullRequestInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	Repo          string `json:"repo" jsonschema:"repository name or ID"`
	PullRequestID int    `json:"pullRequestId" jsonschema:"pull request ID"`
	Status        string `json:"status,omitempty" jsonschema:"new status: completed, abandoned, or active (optional)"`
	Title         string `json:"title,omitempty" jsonschema:"new title (optional)"`
	Description   string `json:"description,omitempty" jsonschema:"new description (optional)"`
}

// AddPRCommentInput adds a comment.
type AddPRCommentInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	Repo          string `json:"repo" jsonschema:"repository name or ID"`
	PullRequestID int    `json:"pullRequestId" jsonschema:"pull request ID"`
	Content       string `json:"content" jsonschema:"comment text"`
}

// VoteInput sets a reviewer vote.
type VoteInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	Repo          string `json:"repo" jsonschema:"repository name or ID"`
	PullRequestID int    `json:"pullRequestId" jsonschema:"pull request ID"`
	ReviewerID    string `json:"reviewerId" jsonschema:"reviewer identity ID"`
	Vote          int    `json:"vote" jsonschema:"vote: 10 approve, 5 approve with suggestions, 0 reset, -5 waiting, -10 reject"`
}

// CreateRepositoryInput describes a new repository.
type CreateRepositoryInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Name    string `json:"name" jsonschema:"new repository name"`
}

// DeleteRepositoryInput identifies a repository to delete.
type DeleteRepositoryInput struct {
	RepoID string `json:"repoId" jsonschema:"repository ID to delete"`
}

// ListCommitStatusesInput identifies a commit whose statuses to list.
type ListCommitStatusesInput struct {
	Project  string `json:"project" jsonschema:"project name or ID"`
	Repo     string `json:"repo" jsonschema:"repository name or ID"`
	CommitID string `json:"commitId" jsonschema:"commit SHA"`
}

// CreateCommitStatusInput describes a new commit status.
type CreateCommitStatusInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	Repo         string `json:"repo" jsonschema:"repository name or ID"`
	CommitID     string `json:"commitId" jsonschema:"commit SHA"`
	State        string `json:"state" jsonschema:"status state: succeeded, failed, pending, error, or notApplicable"`
	Description  string `json:"description,omitempty" jsonschema:"status description (optional)"`
	TargetURL    string `json:"targetUrl,omitempty" jsonschema:"URL with more details (optional)"`
	ContextName  string `json:"contextName" jsonschema:"status context name"`
	ContextGenre string `json:"contextGenre,omitempty" jsonschema:"status context genre (optional)"`
}

// CreateBranchInput describes a new branch.
type CreateBranchInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Repo    string `json:"repo" jsonschema:"repository name or ID"`
	Branch  string `json:"branch" jsonschema:"new branch name, e.g. feature/x (without refs/heads/)"`
	SHA     string `json:"sha" jsonschema:"commit SHA the new branch should point at"`
}

// --- Tool handlers ---

func (s *service) listRepositories(ctx context.Context, _ *mcp.CallToolRequest, in ProjectInput) (*mcp.CallToolResult, server.ListResult[Repository], error) {
	out, err := s.ListRepositories(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) getRepository(ctx context.Context, _ *mcp.CallToolRequest, in RepoInput) (*mcp.CallToolResult, *Repository, error) {
	out, err := s.GetRepository(ctx, in.Project, in.Repo)
	return nil, out, err
}

func (s *service) listRefs(ctx context.Context, _ *mcp.CallToolRequest, in ListRefsInput) (*mcp.CallToolResult, server.ListResult[Ref], error) {
	out, err := s.ListRefs(ctx, in.Project, in.Repo, in.Filter)
	return nil, server.List(out), err
}

func (s *service) listCommits(ctx context.Context, _ *mcp.CallToolRequest, in ListCommitsInput) (*mcp.CallToolResult, server.ListResult[Commit], error) {
	out, err := s.ListCommits(ctx, in.Project, in.Repo, in.Branch, in.Top)
	return nil, server.List(out), err
}

func (s *service) getFile(ctx context.Context, _ *mcp.CallToolRequest, in GetFileInput) (*mcp.CallToolResult, *Item, error) {
	out, err := s.GetFile(ctx, in.Project, in.Repo, in.Path, in.Branch)
	return nil, out, err
}

func (s *service) listPullRequests(ctx context.Context, _ *mcp.CallToolRequest, in ListPullRequestsInput) (*mcp.CallToolResult, server.ListResult[PullRequest], error) {
	out, err := s.ListPullRequests(ctx, in.Project, in.Repo, in.Status, in.Top)
	return nil, server.List(out), err
}

func (s *service) getPullRequest(ctx context.Context, _ *mcp.CallToolRequest, in PullRequestInput) (*mcp.CallToolResult, *PullRequest, error) {
	out, err := s.GetPullRequest(ctx, in.Project, in.Repo, in.PullRequestID)
	return nil, out, err
}

func (s *service) listThreads(ctx context.Context, _ *mcp.CallToolRequest, in PullRequestInput) (*mcp.CallToolResult, server.ListResult[Thread], error) {
	out, err := s.ListThreads(ctx, in.Project, in.Repo, in.PullRequestID)
	return nil, server.List(out), err
}

func (s *service) createPullRequest(ctx context.Context, _ *mcp.CallToolRequest, in CreatePullRequestInput) (*mcp.CallToolResult, *PullRequest, error) {
	body := map[string]any{
		"sourceRefName": normalizeRef(in.SourceBranch),
		"targetRefName": normalizeRef(in.TargetBranch),
		"title":         in.Title,
	}
	if in.Description != "" {
		body["description"] = in.Description
	}
	if in.IsDraft {
		body["isDraft"] = true
	}
	out, err := s.CreatePullRequest(ctx, in.Project, in.Repo, body)
	return nil, out, err
}

func (s *service) updatePullRequest(ctx context.Context, _ *mcp.CallToolRequest, in UpdatePullRequestInput) (*mcp.CallToolResult, *PullRequest, error) {
	body := map[string]any{}
	if in.Status != "" {
		body["status"] = in.Status
	}
	if in.Title != "" {
		body["title"] = in.Title
	}
	if in.Description != "" {
		body["description"] = in.Description
	}
	out, err := s.UpdatePullRequest(ctx, in.Project, in.Repo, in.PullRequestID, body)
	return nil, out, err
}

func (s *service) addComment(ctx context.Context, _ *mcp.CallToolRequest, in AddPRCommentInput) (*mcp.CallToolResult, *Thread, error) {
	out, err := s.AddComment(ctx, in.Project, in.Repo, in.PullRequestID, in.Content)
	return nil, out, err
}

func (s *service) vote(ctx context.Context, _ *mcp.CallToolRequest, in VoteInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.SetVote(ctx, in.Project, in.Repo, in.PullRequestID, in.ReviewerID, in.Vote); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

func (s *service) listCommitStatuses(ctx context.Context, _ *mcp.CallToolRequest, in ListCommitStatusesInput) (*mcp.CallToolResult, server.ListResult[CommitStatus], error) {
	out, err := s.ListCommitStatuses(ctx, in.Project, in.Repo, in.CommitID)
	return nil, server.List(out), err
}

func (s *service) createRepository(ctx context.Context, _ *mcp.CallToolRequest, in CreateRepositoryInput) (*mcp.CallToolResult, *Repository, error) {
	out, err := s.CreateRepository(ctx, in.Project, in.Name)
	return nil, out, err
}

func (s *service) deleteRepository(ctx context.Context, _ *mcp.CallToolRequest, in DeleteRepositoryInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.DeleteRepository(ctx, in.RepoID); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

func (s *service) createCommitStatus(ctx context.Context, _ *mcp.CallToolRequest, in CreateCommitStatusInput) (*mcp.CallToolResult, *CommitStatus, error) {
	contextBody := map[string]any{"name": in.ContextName}
	if in.ContextGenre != "" {
		contextBody["genre"] = in.ContextGenre
	}
	body := map[string]any{
		"state":   in.State,
		"context": contextBody,
	}
	if in.Description != "" {
		body["description"] = in.Description
	}
	if in.TargetURL != "" {
		body["targetUrl"] = in.TargetURL
	}
	out, err := s.CreateCommitStatus(ctx, in.Project, in.Repo, in.CommitID, body)
	return nil, out, err
}

func (s *service) createBranch(ctx context.Context, _ *mcp.CallToolRequest, in CreateBranchInput) (*mcp.CallToolResult, *UpdateRefResult, error) {
	out, err := s.CreateBranch(ctx, in.Project, in.Repo, in.Branch, in.SHA)
	return nil, out, err
}

// normalizeRef ensures a branch is a full ref name (refs/heads/...).
func normalizeRef(branch string) string {
	if branch == "" {
		return ""
	}
	if len(branch) >= 5 && branch[:5] == "refs/" {
		return branch
	}
	return "refs/heads/" + branch
}

// --- Reviewer tools ---

// AddReviewerInput adds or updates a reviewer on a pull request.
type AddReviewerInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	Repo          string `json:"repo" jsonschema:"repository name or ID"`
	PullRequestID int    `json:"pullRequestId" jsonschema:"pull request ID"`
	ReviewerID    string `json:"reviewerId" jsonschema:"reviewer identity ID"`
	Required      bool   `json:"required,omitempty" jsonschema:"mark the reviewer as required (optional)"`
}

// RemoveReviewerInput removes a reviewer from a pull request.
type RemoveReviewerInput struct {
	Project       string `json:"project" jsonschema:"project name or ID"`
	Repo          string `json:"repo" jsonschema:"repository name or ID"`
	PullRequestID int    `json:"pullRequestId" jsonschema:"pull request ID"`
	ReviewerID    string `json:"reviewerId" jsonschema:"reviewer identity ID"`
}

func (s *service) listReviewers(ctx context.Context, _ *mcp.CallToolRequest, in PullRequestInput) (*mcp.CallToolResult, server.ListResult[Reviewer], error) {
	out, err := s.ListReviewers(ctx, in.Project, in.Repo, in.PullRequestID)
	return nil, server.List(out), err
}

func (s *service) addReviewer(ctx context.Context, _ *mcp.CallToolRequest, in AddReviewerInput) (*mcp.CallToolResult, *Reviewer, error) {
	out, err := s.AddReviewer(ctx, in.Project, in.Repo, in.PullRequestID, in.ReviewerID, in.Required)
	return nil, out, err
}

func (s *service) removeReviewer(ctx context.Context, _ *mcp.CallToolRequest, in RemoveReviewerInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.RemoveReviewer(ctx, in.Project, in.Repo, in.PullRequestID, in.ReviewerID); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

// PushFileInput commits a single file change to a branch.
type PushFileInput struct {
	Project    string `json:"project" jsonschema:"project name or ID"`
	Repo       string `json:"repo" jsonschema:"repository name or ID"`
	Branch     string `json:"branch" jsonschema:"branch name, e.g. main"`
	Path       string `json:"path" jsonschema:"file path, e.g. /src/main.go"`
	Content    string `json:"content,omitempty" jsonschema:"new file content (omit for delete)"`
	Message    string `json:"message" jsonschema:"commit message"`
	ChangeType string `json:"changeType,omitempty" jsonschema:"add (default), edit, or delete"`
}

func (s *service) pushFile(ctx context.Context, _ *mcp.CallToolRequest, in PushFileInput) (*mcp.CallToolResult, *Push, error) {
	out, err := s.PushFile(ctx, in.Project, in.Repo, in.Branch, in.Path, in.Content, in.Message, in.ChangeType)
	return nil, out, err
}
