// SPDX-License-Identifier: MIT

//go:build integration

package git

import (
	"context"
	"os"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

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

// firstProject finds a project to scope repository queries to. It honors
// ADO_TEST_PROJECT, otherwise picks the first project in the org.
func firstProject(t *testing.T, c *ado.Clients) string {
	t.Helper()
	if p := os.Getenv("ADO_TEST_PROJECT"); p != "" {
		return p
	}
	var out client.List[struct {
		Name string `json:"name"`
	}]
	if err := c.Org.GetJSON(context.Background(), "/_apis/projects", nil, &out); err != nil {
		t.Fatalf("listing projects: %v", err)
	}
	if len(out.Value) == 0 {
		t.Skip("no projects in org")
	}
	return out.Value[0].Name
}

func TestIntegration_ListRepositories(t *testing.T) {
	c := integrationClients(t)
	svc := &service{c: c}
	project := firstProject(t, c)

	repos, err := svc.ListRepositories(context.Background(), project)
	if err != nil {
		t.Fatalf("ListRepositories(%q): %v", project, err)
	}
	t.Logf("project %q has %d repositories", project, len(repos))
	for _, r := range repos {
		if r.Name == "" {
			t.Errorf("repository with empty name: %+v", r)
		}
	}
}
