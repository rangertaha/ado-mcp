// SPDX-License-Identifier: MIT

package pipelines

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Pipelines/Build toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "pipelines_list", Title: "List pipelines",
		Description: "List the pipelines in a project."}, svc.listPipelines)
	server.Register(s, server.ToolDef{Name: "pipelines_get", Title: "Get pipeline",
		Description: "Get a pipeline by ID."}, svc.getPipeline)
	server.Register(s, server.ToolDef{Name: "pipelines_list_runs", Title: "List pipeline runs",
		Description: "List the runs of a pipeline."}, svc.listRuns)
	server.Register(s, server.ToolDef{Name: "pipelines_get_run", Title: "Get pipeline run",
		Description: "Get a single pipeline run by ID."}, svc.getRun)
	server.Register(s, server.ToolDef{Name: "pipelines_list_builds", Title: "List builds",
		Description: "List builds in a project, optionally filtered by build definition."}, svc.listBuilds)
	server.Register(s, server.ToolDef{Name: "pipelines_get_build", Title: "Get build",
		Description: "Get a single build by ID."}, svc.getBuild)
	server.Register(s, server.ToolDef{Name: "pipelines_get_build_log", Title: "Get build log",
		Description: "Get the text of a specific build log."}, svc.getBuildLog)
	server.Register(s, server.ToolDef{Name: "pipelines_list_definitions", Title: "List build definitions",
		Description: "List the build definitions in a project."}, svc.listDefinitions)
	server.Register(s, server.ToolDef{Name: "pipelines_list_artifacts", Title: "List build artifacts",
		Description: "List the artifacts produced by a build."}, svc.listArtifacts)
	server.Register(s, server.ToolDef{Name: "pipelines_list_build_logs", Title: "List build logs",
		Description: "List the logs produced by a build."}, svc.listBuildLogs)
	server.Register(s, server.ToolDef{Name: "pipelines_get_build_timeline", Title: "Get build timeline",
		Description: "Get the execution timeline of a build."}, svc.getBuildTimeline)
	server.Register(s, server.ToolDef{Name: "pipelines_get_build_changes", Title: "Get build changes",
		Description: "Get the source changes associated with a build."}, svc.getBuildChanges)

	// --- Write tools ---
	server.Register(s, server.ToolDef{Name: "pipelines_run", Title: "Run pipeline",
		Description: "Queue a new run of a pipeline, optionally on a specific branch.", Write: true}, svc.runPipeline)
	server.Register(s, server.ToolDef{Name: "pipelines_cancel_build", Title: "Cancel build",
		Description: "Request cancellation of an in-progress build.", Write: true, Idempotent: true}, svc.cancelBuild)
}

// --- Tool input types ---

// ProjectInput identifies a project.
type ProjectInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

// ListPipelinesInput pages pipelines.
type ListPipelinesInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Top     int    `json:"top,omitempty" jsonschema:"maximum number of pipelines (optional)"`
}

// PipelineInput identifies a pipeline.
type PipelineInput struct {
	Project    string `json:"project" jsonschema:"project name or ID"`
	PipelineID int    `json:"pipelineId" jsonschema:"pipeline ID"`
}

// RunInput identifies a pipeline run.
type RunInput struct {
	Project    string `json:"project" jsonschema:"project name or ID"`
	PipelineID int    `json:"pipelineId" jsonschema:"pipeline ID"`
	RunID      int    `json:"runId" jsonschema:"run ID"`
}

// RunPipelineInput queues a pipeline run.
type RunPipelineInput struct {
	Project    string `json:"project" jsonschema:"project name or ID"`
	PipelineID int    `json:"pipelineId" jsonschema:"pipeline ID"`
	Branch     string `json:"branch,omitempty" jsonschema:"branch to run on (optional, defaults to the default branch)"`
}

// ListBuildsInput filters builds.
type ListBuildsInput struct {
	Project      string `json:"project" jsonschema:"project name or ID"`
	DefinitionID int    `json:"definitionId,omitempty" jsonschema:"restrict to a build definition (optional)"`
	Top          int    `json:"top,omitempty" jsonschema:"maximum number of builds (optional)"`
}

// BuildInput identifies a build.
type BuildInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	BuildID int    `json:"buildId" jsonschema:"build ID"`
}

// BuildLogInput identifies a build log.
type BuildLogInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	BuildID int    `json:"buildId" jsonschema:"build ID"`
	LogID   int    `json:"logId" jsonschema:"log ID"`
}

// BuildLogOutput wraps build log text (structured output must be an object).
type BuildLogOutput struct {
	Log string `json:"log" jsonschema:"the build log text"`
}

// --- Tool handlers ---

func (s *service) listPipelines(ctx context.Context, _ *mcp.CallToolRequest, in ListPipelinesInput) (*mcp.CallToolResult, server.ListResult[Pipeline], error) {
	out, err := s.ListPipelines(ctx, in.Project, in.Top)
	return nil, server.List(out), err
}

func (s *service) getPipeline(ctx context.Context, _ *mcp.CallToolRequest, in PipelineInput) (*mcp.CallToolResult, *Pipeline, error) {
	out, err := s.GetPipeline(ctx, in.Project, in.PipelineID)
	return nil, out, err
}

func (s *service) listRuns(ctx context.Context, _ *mcp.CallToolRequest, in PipelineInput) (*mcp.CallToolResult, server.ListResult[Run], error) {
	out, err := s.ListRuns(ctx, in.Project, in.PipelineID)
	return nil, server.List(out), err
}

func (s *service) getRun(ctx context.Context, _ *mcp.CallToolRequest, in RunInput) (*mcp.CallToolResult, *Run, error) {
	out, err := s.GetRun(ctx, in.Project, in.PipelineID, in.RunID)
	return nil, out, err
}

func (s *service) listBuilds(ctx context.Context, _ *mcp.CallToolRequest, in ListBuildsInput) (*mcp.CallToolResult, server.ListResult[Build], error) {
	out, err := s.ListBuilds(ctx, in.Project, in.DefinitionID, in.Top)
	return nil, server.List(out), err
}

func (s *service) getBuild(ctx context.Context, _ *mcp.CallToolRequest, in BuildInput) (*mcp.CallToolResult, *Build, error) {
	out, err := s.GetBuild(ctx, in.Project, in.BuildID)
	return nil, out, err
}

func (s *service) getBuildLog(ctx context.Context, _ *mcp.CallToolRequest, in BuildLogInput) (*mcp.CallToolResult, *BuildLogOutput, error) {
	out, err := s.GetBuildLog(ctx, in.Project, in.BuildID, in.LogID)
	if err != nil {
		return nil, nil, err
	}
	return nil, &BuildLogOutput{Log: out}, nil
}

func (s *service) listDefinitions(ctx context.Context, _ *mcp.CallToolRequest, in ProjectInput) (*mcp.CallToolResult, server.ListResult[Definition], error) {
	out, err := s.ListDefinitions(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) listArtifacts(ctx context.Context, _ *mcp.CallToolRequest, in BuildInput) (*mcp.CallToolResult, server.ListResult[Artifact], error) {
	out, err := s.ListArtifacts(ctx, in.Project, in.BuildID)
	return nil, server.List(out), err
}

func (s *service) runPipeline(ctx context.Context, _ *mcp.CallToolRequest, in RunPipelineInput) (*mcp.CallToolResult, *Run, error) {
	out, err := s.RunPipeline(ctx, in.Project, in.PipelineID, in.Branch)
	return nil, out, err
}

func (s *service) listBuildLogs(ctx context.Context, _ *mcp.CallToolRequest, in BuildInput) (*mcp.CallToolResult, server.ListResult[BuildLog], error) {
	out, err := s.ListBuildLogs(ctx, in.Project, in.BuildID)
	return nil, server.List(out), err
}

func (s *service) getBuildTimeline(ctx context.Context, _ *mcp.CallToolRequest, in BuildInput) (*mcp.CallToolResult, *Timeline, error) {
	out, err := s.GetBuildTimeline(ctx, in.Project, in.BuildID)
	return nil, out, err
}

func (s *service) getBuildChanges(ctx context.Context, _ *mcp.CallToolRequest, in BuildInput) (*mcp.CallToolResult, server.ListResult[Change], error) {
	out, err := s.GetBuildChanges(ctx, in.Project, in.BuildID)
	return nil, server.List(out), err
}

func (s *service) cancelBuild(ctx context.Context, _ *mcp.CallToolRequest, in BuildInput) (*mcp.CallToolResult, *Build, error) {
	out, err := s.CancelBuild(ctx, in.Project, in.BuildID)
	return nil, out, err
}
