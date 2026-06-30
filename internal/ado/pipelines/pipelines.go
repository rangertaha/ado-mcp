// SPDX-License-Identifier: MIT

// Package pipelines exposes the Azure DevOps Pipelines and Build services:
// pipelines, pipeline runs, build definitions, builds, build logs and artifacts.
package pipelines

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "pipelines"

// service wraps the Azure DevOps clients for Pipelines/Build operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Pipeline is a YAML/classic pipeline definition.
type Pipeline struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Folder   string `json:"folder,omitempty"`
	Revision int    `json:"revision,omitempty"`
	URL      string `json:"url,omitempty"`
}

// Run is a pipeline run.
type Run struct {
	ID           int    `json:"id"`
	Name         string `json:"name,omitempty"`
	State        string `json:"state,omitempty"`
	Result       string `json:"result,omitempty"`
	CreatedDate  string `json:"createdDate,omitempty"`
	FinishedDate string `json:"finishedDate,omitempty"`
	URL          string `json:"url,omitempty"`
}

// Build is a build (the Build service view of a run).
type Build struct {
	ID            int    `json:"id"`
	BuildNumber   string `json:"buildNumber,omitempty"`
	Status        string `json:"status,omitempty"`
	Result        string `json:"result,omitempty"`
	SourceBranch  string `json:"sourceBranch,omitempty"`
	SourceVersion string `json:"sourceVersion,omitempty"`
	Definition    any    `json:"definition,omitempty"`
	URL           string `json:"url,omitempty"`
}

// Definition is a build definition.
type Definition struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	Revision int    `json:"revision,omitempty"`
	URL      string `json:"url,omitempty"`
}

// Artifact is a build artifact.
type Artifact struct {
	ID       int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Resource any    `json:"resource,omitempty"`
}

// BuildLog is a single log produced by a build.
type BuildLog struct {
	ID        int    `json:"id"`
	Type      string `json:"type,omitempty"`
	LineCount int    `json:"lineCount,omitempty"`
	CreatedOn string `json:"createdOn,omitempty"`
	URL       string `json:"url,omitempty"`
}

// Timeline is the execution timeline of a build.
type Timeline struct {
	ID      string `json:"id"`
	Records any    `json:"records,omitempty"`
}

// Change is a source change associated with a build.
type Change struct {
	ID      string `json:"id"`
	Message string `json:"message,omitempty"`
	Type    string `json:"type,omitempty"`
	Author  any    `json:"author,omitempty"`
}

// --- Pipeline operations ---

// ListPipelines returns the pipelines in a project.
func (s *service) ListPipelines(ctx context.Context, project string, top int) ([]Pipeline, error) {
	q := url.Values{}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[Pipeline]
	path := fmt.Sprintf("/%s/_apis/pipelines", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetPipeline returns a single pipeline by ID.
func (s *service) GetPipeline(ctx context.Context, project string, id int) (*Pipeline, error) {
	var p Pipeline
	path := fmt.Sprintf("/%s/_apis/pipelines/%d", url.PathEscape(project), id)
	if err := s.c.Org.GetJSON(ctx, path, nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// ListRuns returns the runs of a pipeline.
func (s *service) ListRuns(ctx context.Context, project string, pipelineID int) ([]Run, error) {
	var out client.List[Run]
	path := fmt.Sprintf("/%s/_apis/pipelines/%d/runs", url.PathEscape(project), pipelineID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetRun returns a single pipeline run.
func (s *service) GetRun(ctx context.Context, project string, pipelineID, runID int) (*Run, error) {
	var r Run
	path := fmt.Sprintf("/%s/_apis/pipelines/%d/runs/%d", url.PathEscape(project), pipelineID, runID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// RunPipeline queues a new run of a pipeline, optionally on a branch.
func (s *service) RunPipeline(ctx context.Context, project string, pipelineID int, branch string) (*Run, error) {
	body := map[string]any{}
	if branch != "" {
		body["resources"] = map[string]any{
			"repositories": map[string]any{
				"self": map[string]any{"refName": normalizeRef(branch)},
			},
		}
	}
	var r Run
	path := fmt.Sprintf("/%s/_apis/pipelines/%d/runs", url.PathEscape(project), pipelineID)
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// --- Build operations ---

// ListBuilds returns builds in a project, optionally filtered by definition.
func (s *service) ListBuilds(ctx context.Context, project string, definitionID, top int) ([]Build, error) {
	q := url.Values{}
	if definitionID > 0 {
		q.Set("definitions", strconv.Itoa(definitionID))
	}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[Build]
	path := fmt.Sprintf("/%s/_apis/build/builds", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetBuild returns a single build by ID.
func (s *service) GetBuild(ctx context.Context, project string, id int) (*Build, error) {
	var b Build
	path := fmt.Sprintf("/%s/_apis/build/builds/%d", url.PathEscape(project), id)
	if err := s.c.Org.GetJSON(ctx, path, nil, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// GetBuildLog returns the text of a specific build log.
func (s *service) GetBuildLog(ctx context.Context, project string, buildID, logID int) (string, error) {
	path := fmt.Sprintf("/%s/_apis/build/builds/%d/logs/%d", url.PathEscape(project), buildID, logID)
	var raw client.RawBody
	if _, err := s.c.Org.Do(ctx, client.Request{Method: "GET", Path: path, Out: &raw}); err != nil {
		return "", err
	}
	return raw.String(), nil
}

// ListDefinitions returns the build definitions in a project.
func (s *service) ListDefinitions(ctx context.Context, project string) ([]Definition, error) {
	var out client.List[Definition]
	path := fmt.Sprintf("/%s/_apis/build/definitions", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListArtifacts returns the artifacts produced by a build.
func (s *service) ListArtifacts(ctx context.Context, project string, buildID int) ([]Artifact, error) {
	var out client.List[Artifact]
	path := fmt.Sprintf("/%s/_apis/build/builds/%d/artifacts", url.PathEscape(project), buildID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListBuildLogs returns the logs produced by a build.
func (s *service) ListBuildLogs(ctx context.Context, project string, buildID int) ([]BuildLog, error) {
	var out client.List[BuildLog]
	path := fmt.Sprintf("/%s/_apis/build/builds/%d/logs", url.PathEscape(project), buildID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetBuildTimeline returns the execution timeline of a build.
func (s *service) GetBuildTimeline(ctx context.Context, project string, buildID int) (*Timeline, error) {
	var t Timeline
	path := fmt.Sprintf("/%s/_apis/build/builds/%d/timeline", url.PathEscape(project), buildID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// GetBuildChanges returns the source changes associated with a build.
func (s *service) GetBuildChanges(ctx context.Context, project string, buildID int) ([]Change, error) {
	var out client.List[Change]
	path := fmt.Sprintf("/%s/_apis/build/builds/%d/changes", url.PathEscape(project), buildID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// CancelBuild requests cancellation of an in-progress build.
func (s *service) CancelBuild(ctx context.Context, project string, buildID int) (*Build, error) {
	body := map[string]any{"status": "cancelling"}
	var b Build
	path := fmt.Sprintf("/%s/_apis/build/builds/%d", url.PathEscape(project), buildID)
	if err := s.c.Org.PatchJSON(ctx, path, nil, body, &b, ""); err != nil {
		return nil, err
	}
	return &b, nil
}

// normalizeRef ensures a branch is a full ref name.
func normalizeRef(branch string) string {
	if branch == "" || len(branch) >= 5 && branch[:5] == "refs/" {
		return branch
	}
	return "refs/heads/" + branch
}
