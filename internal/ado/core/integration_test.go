// SPDX-License-Identifier: MIT

//go:build integration

package core

import (
	"context"
	"os"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/ado"
)

// integrationClients builds real Azure DevOps clients from the environment,
// skipping the test when credentials are absent.
func integrationClients(t *testing.T) *ado.Clients {
	t.Helper()
	org, pat := os.Getenv("ADO_ORG_URL"), os.Getenv("ADO_PAT")
	if org == "" || pat == "" {
		t.Skip("ADO_ORG_URL/ADO_PAT not set; skipping integration test")
	}
	c, err := ado.NewClients(org, pat)
	if err != nil {
		t.Fatalf("NewClients: %v", err)
	}
	return c
}

func TestIntegration_ListProjects(t *testing.T) {
	svc := &service{c: integrationClients(t)}
	projects, err := svc.ListProjects(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(projects) == 0 {
		t.Fatal("expected at least one project")
	}
	t.Logf("found %d projects; first: %q (%s)", len(projects), projects[0].Name, projects[0].State)
	for _, p := range projects {
		if p.Name == "" {
			t.Errorf("project with empty name: %+v", p)
		}
	}
}

func TestIntegration_ListProcesses(t *testing.T) {
	svc := &service{c: integrationClients(t)}
	procs, err := svc.ListProcesses(context.Background())
	if err != nil {
		t.Fatalf("ListProcesses: %v", err)
	}
	t.Logf("found %d processes", len(procs))
}
