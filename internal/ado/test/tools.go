// SPDX-License-Identifier: MIT

package test

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Test Plans toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "test_list_plans", Title: "List test plans",
		Description: "List the test plans in a project."}, svc.listPlans)
	server.Register(s, server.ToolDef{Name: "test_get_plan", Title: "Get test plan",
		Description: "Get a single test plan by ID."}, svc.getPlan)
	server.Register(s, server.ToolDef{Name: "test_list_suites", Title: "List test suites",
		Description: "List the test suites in a test plan."}, svc.listSuites)
	server.Register(s, server.ToolDef{Name: "test_list_cases", Title: "List test cases",
		Description: "List the test cases in a test suite."}, svc.listTestCases)
	server.Register(s, server.ToolDef{Name: "test_list_runs", Title: "List test runs",
		Description: "List the test runs in a project."}, svc.listRuns)
	server.Register(s, server.ToolDef{Name: "test_get_run", Title: "Get test run",
		Description: "Get a single test run by ID."}, svc.getRun)
	server.Register(s, server.ToolDef{Name: "test_list_results", Title: "List test results",
		Description: "List the results for a test run."}, svc.listResults)
	server.Register(s, server.ToolDef{Name: "test_create_run", Title: "Create test run",
		Description: "Create a new manual test run for a test plan.", Write: true}, svc.createRun)
	server.Register(s, server.ToolDef{Name: "test_update_run", Title: "Update test run",
		Description: "Update the state of a test run (e.g. InProgress, Completed, Aborted).", Write: true, Idempotent: true}, svc.updateRun)
	server.Register(s, server.ToolDef{Name: "test_add_results", Title: "Add test results",
		Description: "Add test results to a test run.", Write: true}, svc.addResults)
}

// ListPlansInput selects a project.
type ListPlansInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

func (s *service) listPlans(ctx context.Context, _ *mcp.CallToolRequest, in ListPlansInput) (*mcp.CallToolResult, server.ListResult[TestPlan], error) {
	out, err := s.ListPlans(ctx, in.Project)
	return nil, server.List(out), err
}

// GetPlanInput identifies a test plan.
type GetPlanInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	PlanID  int    `json:"planId" jsonschema:"the test plan ID"`
}

func (s *service) getPlan(ctx context.Context, _ *mcp.CallToolRequest, in GetPlanInput) (*mcp.CallToolResult, *TestPlan, error) {
	out, err := s.GetPlan(ctx, in.Project, in.PlanID)
	return nil, out, err
}

// ListSuitesInput identifies a test plan whose suites are listed.
type ListSuitesInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	PlanID  int    `json:"planId" jsonschema:"the test plan ID"`
}

func (s *service) listSuites(ctx context.Context, _ *mcp.CallToolRequest, in ListSuitesInput) (*mcp.CallToolResult, server.ListResult[TestSuite], error) {
	out, err := s.ListSuites(ctx, in.Project, in.PlanID)
	return nil, server.List(out), err
}

// ListTestCasesInput identifies a test suite whose test cases are listed.
type ListTestCasesInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	PlanID  int    `json:"planId" jsonschema:"the test plan ID"`
	SuiteID int    `json:"suiteId" jsonschema:"the test suite ID"`
}

func (s *service) listTestCases(ctx context.Context, _ *mcp.CallToolRequest, in ListTestCasesInput) (*mcp.CallToolResult, server.ListResult[TestCase], error) {
	out, err := s.ListTestCases(ctx, in.Project, in.PlanID, in.SuiteID)
	return nil, server.List(out), err
}

// ListRunsInput selects a project and optionally limits the number of runs.
type ListRunsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Top     int    `json:"top,omitempty" jsonschema:"maximum number of test runs (optional)"`
}

func (s *service) listRuns(ctx context.Context, _ *mcp.CallToolRequest, in ListRunsInput) (*mcp.CallToolResult, server.ListResult[TestRun], error) {
	out, err := s.ListRuns(ctx, in.Project, in.Top)
	return nil, server.List(out), err
}

// GetRunInput identifies a test run.
type GetRunInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	RunID   int    `json:"runId" jsonschema:"the test run ID"`
}

func (s *service) getRun(ctx context.Context, _ *mcp.CallToolRequest, in GetRunInput) (*mcp.CallToolResult, *TestRun, error) {
	out, err := s.GetRun(ctx, in.Project, in.RunID)
	return nil, out, err
}

// ListResultsInput identifies a test run whose results are listed.
type ListResultsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	RunID   int    `json:"runId" jsonschema:"the test run ID"`
}

func (s *service) listResults(ctx context.Context, _ *mcp.CallToolRequest, in ListResultsInput) (*mcp.CallToolResult, server.ListResult[TestResult], error) {
	out, err := s.ListResults(ctx, in.Project, in.RunID)
	return nil, server.List(out), err
}

// CreateRunInput specifies the test run to create.
type CreateRunInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Name    string `json:"name" jsonschema:"the name of the test run"`
	PlanID  int    `json:"planId" jsonschema:"the test plan ID the run belongs to"`
}

func (s *service) createRun(ctx context.Context, _ *mcp.CallToolRequest, in CreateRunInput) (*mcp.CallToolResult, *TestRun, error) {
	out, err := s.CreateRun(ctx, in.Project, in.Name, in.PlanID)
	return nil, out, err
}

// UpdateRunInput specifies the test run and its new state.
type UpdateRunInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	RunID   int    `json:"runId" jsonschema:"the test run ID"`
	State   string `json:"state" jsonschema:"the new state (e.g. InProgress, Completed, Aborted)"`
}

func (s *service) updateRun(ctx context.Context, _ *mcp.CallToolRequest, in UpdateRunInput) (*mcp.CallToolResult, *TestRun, error) {
	out, err := s.UpdateRun(ctx, in.Project, in.RunID, in.State)
	return nil, out, err
}

// AddResultsInput specifies the test run and the results to add.
type AddResultsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	RunID   int    `json:"runId" jsonschema:"the test run ID"`
	Results []any  `json:"results" jsonschema:"the test results to add (array of result objects)"`
}

func (s *service) addResults(ctx context.Context, _ *mcp.CallToolRequest, in AddResultsInput) (*mcp.CallToolResult, server.ListResult[any], error) {
	out, err := s.AddResults(ctx, in.Project, in.RunID, in.Results)
	return nil, server.List(out), err
}
