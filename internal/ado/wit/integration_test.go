// SPDX-License-Identifier: MIT

//go:build integration

package wit

import (
	"context"
	"os"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/ado"
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

// ListFields is organization-level, so it needs no project.
func TestIntegration_ListFields(t *testing.T) {
	svc := &service{c: integrationClients(t)}
	fields, err := svc.ListFields(context.Background())
	if err != nil {
		t.Fatalf("ListFields: %v", err)
	}
	if len(fields) == 0 {
		t.Fatal("expected at least one work item field")
	}
	t.Logf("found %d fields; e.g. %q", len(fields), fields[0].ReferenceName)
}
