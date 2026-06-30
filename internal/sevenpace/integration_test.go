// SPDX-License-Identifier: MIT

//go:build integration

package sevenpace

import (
	"context"
	"strings"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/config"
)

// integrationClient builds a real 7pace client from the resolved configuration,
// skipping when 7pace is not configured.
func integrationClient(t *testing.T) *Client {
	t.Helper()
	cfg, err := config.Load()
	if err != nil || !cfg.SevenPaceEnabled() {
		t.Skip("7pace not configured; skipping integration test")
	}
	c, err := NewClient(cfg.SevenPaceBaseURL(), cfg.SevenPaceToken)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Logf("7pace base: %s", cfg.SevenPaceBaseURL())
	return c
}

// The service document at the root lists the entity sets; fetching it verifies
// the base URL, auth, and connectivity.
func TestIntegration_ServiceRoot(t *testing.T) {
	c := integrationClient(t)
	body, err := c.QueryRaw(context.Background(), "", nil)
	if err != nil {
		t.Fatalf("QueryRaw(service root): %v", err)
	}
	if !strings.Contains(body, "workLogsOnly") {
		t.Fatalf("service document missing expected entity sets: %.300s", body)
	}
	t.Logf("service root OK: %.300s", body)
}

// Reading worklogs requires a Service Account to be configured in 7pace
// Timetracker settings; when it isn't, the API returns 403 "ServiceAccountNotSet".
// We treat that as a skip (tenant configuration), not a failure.
func TestIntegration_WorkLogs(t *testing.T) {
	c := integrationClient(t)
	logs, err := c.ListWorkLogs(context.Background(), "", 5)
	if err != nil {
		if strings.Contains(err.Error(), "ServiceAccountNotSet") {
			t.Skipf("7pace Service Account not configured in tenant settings: %v", err)
		}
		t.Fatalf("ListWorkLogs: %v", err)
	}
	t.Logf("fetched %d worklogs (top=5)", len(logs))
}
