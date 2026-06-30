// SPDX-License-Identifier: MIT

// Package test provides read-only access to Azure DevOps Test Plans,
// test suites, test cases, test runs, and test results.
package test

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name for the Test Plans area.
const Name = "test"

// service holds the Azure DevOps clients used by the Test Plans area.
type service struct{ c *ado.Clients }

// TestPlan represents an Azure DevOps test plan.
type TestPlan struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AreaPath  string `json:"areaPath,omitempty"`
	Iteration string `json:"iteration,omitempty"`
	State     string `json:"state,omitempty"`
}

// TestSuite represents a test suite within a test plan.
type TestSuite struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	SuiteType   string `json:"suiteType,omitempty"`
	ParentSuite any    `json:"parentSuite,omitempty"`
}

// TestCase represents a test case within a test suite. The work item shape
// varies, so it is kept as an opaque value.
type TestCase struct {
	WorkItem any `json:"workItem,omitempty"`
}

// TestRun represents a test run.
type TestRun struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	State         string `json:"state,omitempty"`
	StartedDate   string `json:"startedDate,omitempty"`
	CompletedDate string `json:"completedDate,omitempty"`
}

// TestResult represents a single result within a test run.
type TestResult struct {
	ID            int    `json:"id"`
	Outcome       string `json:"outcome,omitempty"`
	TestCaseTitle string `json:"testCaseTitle,omitempty"`
	DurationInMs  any    `json:"durationInMs,omitempty"`
}

// ListPlans lists the test plans in a project.
func (s *service) ListPlans(ctx context.Context, project string) ([]TestPlan, error) {
	var out client.List[TestPlan]
	path := fmt.Sprintf("/%s/_apis/testplan/plans", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetPlan gets a single test plan by ID.
func (s *service) GetPlan(ctx context.Context, project string, planID int) (*TestPlan, error) {
	var out TestPlan
	path := fmt.Sprintf("/%s/_apis/testplan/plans/%d", url.PathEscape(project), planID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSuites lists the test suites in a test plan.
func (s *service) ListSuites(ctx context.Context, project string, planID int) ([]TestSuite, error) {
	var out client.List[TestSuite]
	path := fmt.Sprintf("/%s/_apis/testplan/Plans/%d/suites", url.PathEscape(project), planID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListTestCases lists the test cases in a test suite.
func (s *service) ListTestCases(ctx context.Context, project string, planID, suiteID int) ([]TestCase, error) {
	var out client.List[TestCase]
	path := fmt.Sprintf("/%s/_apis/testplan/Plans/%d/Suites/%d/TestCase", url.PathEscape(project), planID, suiteID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListRuns lists the test runs in a project.
func (s *service) ListRuns(ctx context.Context, project string, top int) ([]TestRun, error) {
	q := url.Values{}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	var out client.List[TestRun]
	path := fmt.Sprintf("/%s/_apis/test/runs", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetRun gets a single test run by ID.
func (s *service) GetRun(ctx context.Context, project string, runID int) (*TestRun, error) {
	var out TestRun
	path := fmt.Sprintf("/%s/_apis/test/runs/%d", url.PathEscape(project), runID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateRun creates a new manual test run for a test plan.
func (s *service) CreateRun(ctx context.Context, project, name string, planID int) (*TestRun, error) {
	body := map[string]any{
		"name":      name,
		"plan":      map[string]any{"id": planID},
		"automated": false,
	}
	var out TestRun
	path := fmt.Sprintf("/%s/_apis/test/runs", url.PathEscape(project))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateRun updates the state of a test run (e.g. "InProgress", "Completed", "Aborted").
func (s *service) UpdateRun(ctx context.Context, project string, runID int, state string) (*TestRun, error) {
	body := map[string]any{"state": state}
	var out TestRun
	path := fmt.Sprintf("/%s/_apis/test/runs/%d", url.PathEscape(project), runID)
	if err := s.c.Org.PatchJSON(ctx, path, nil, body, &out, ""); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddResults adds test results to a test run. The results are passed through as-is.
func (s *service) AddResults(ctx context.Context, project string, runID int, results []any) ([]any, error) {
	var out client.List[any]
	path := fmt.Sprintf("/%s/_apis/test/Runs/%d/results", url.PathEscape(project), runID)
	if err := s.c.Org.PostJSON(ctx, path, nil, results, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListResults lists the results for a test run.
func (s *service) ListResults(ctx context.Context, project string, runID int) ([]TestResult, error) {
	var out client.List[TestResult]
	path := fmt.Sprintf("/%s/_apis/test/Runs/%d/results", url.PathEscape(project), runID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
