// SPDX-License-Identifier: MIT

//go:build integration

package main

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/app"
	"github.com/rangertaha/ado-mcp/internal/config"
)

// toleratedErrors are substrings of error messages that reflect tenant
// configuration (not a server bug); tools returning them are reported but do
// not fail the sweep.
var toleratedErrors = []string{
	"ServiceAccountNotSet", // 7pace reporting needs a service account
	"Advanced Security",    // GitHub Advanced Security not enabled
	"advanced security",    //
	"is not enabled",       //
	"TF400813",             // caller lacks permission for a resource
	"social descriptors",   // identities_read requires a search filter (no list-all)
}

// session connects an in-process client to the assembled server.
func session(t *testing.T) (*mcp.ClientSession, func()) {
	t.Helper()
	cfg, err := config.Load()
	if err != nil || cfg.OrgURL == "" || cfg.PAT == "" {
		t.Skip("ADO_ORG_URL/ADO_PAT not set; skipping live integration sweep")
	}
	srv, cleanup, err := app.Assemble(cfg, "test")
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}

	ctx := context.Background()
	clientT, serverT := mcp.NewInMemoryTransports()
	if _, err := srv.Connect(ctx, serverT); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: "sweep", Version: "0"}, nil)
	cs, err := c.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	return cs, func() { _ = cs.Close(); cleanup() }
}

func callTool(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s) protocol error: %v", name, err)
	}
	return res
}

// firstItemField calls a list tool and returns the named string field of its
// first item, or "" if none.
func firstItemField(t *testing.T, cs *mcp.ClientSession, tool, field string, args map[string]any) string {
	res := callTool(t, cs, tool, args)
	if res.IsError {
		return ""
	}
	b, _ := json.Marshal(res.StructuredContent)
	var wrap struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(b, &wrap)
	if len(wrap.Items) == 0 {
		return ""
	}
	if v, ok := wrap.Items[0][field].(string); ok {
		return v
	}
	return ""
}

func errText(res *mcp.CallToolResult) string {
	b, _ := json.Marshal(res.Content)
	return string(b)
}

func tolerated(msg string) bool {
	for _, s := range toleratedErrors {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// TestIntegration_ReadOnlySweep calls every read-only tool whose required
// arguments can be satisfied from discovered context, asserting each succeeds.
func TestIntegration_ReadOnlySweep(t *testing.T) {
	cs, done := session(t)
	defer done()

	ctx := context.Background()
	tools, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	t.Logf("server exposes %d tools", len(tools.Tools))

	// Discover context to feed tool arguments.
	project := firstItemField(t, cs, "core_list_projects", "name", map[string]any{})
	if project == "" {
		t.Fatal("could not discover a project; cannot run sweep")
	}
	team := firstItemField(t, cs, "core_list_teams", "name", map[string]any{"project": project})
	repo := firstItemField(t, cs, "git_list_repositories", "name", map[string]any{"project": project})
	t.Logf("discovered project=%q team=%q repo=%q", project, team, repo)

	// Known fillable argument values, keyed by input field name.
	fill := map[string]string{
		"project":        project,
		"team":           team,
		"repo":           repo,
		"repository":     repo,
		"wiql":           "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project AND [System.ChangedDate] >= @Today-7",
		"structureGroup": "areas",
		"searchText":     "ado",
	}

	// Tools that scan the entire project/org are too expensive for a sweep.
	heavy := map[string]bool{
		"stats_org_contributors":     true,
		"stats_project_contributors": true,
	}

	var ran, ok, skipped, tol int
	var failures []string

	// Stable order for readable output.
	list := append([]*mcp.Tool(nil), tools.Tools...)
	sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })

	for _, tool := range list {
		if a := tool.Annotations; a == nil || !a.ReadOnlyHint {
			continue // read-only sweep
		}
		if heavy[tool.Name] {
			skipped++
			t.Logf("SKIP  %-34s (whole-org/project scan; too slow for sweep)", tool.Name)
			continue
		}
		// Determine required fields from the input schema.
		schemaBytes, _ := json.Marshal(tool.InputSchema)
		var schema struct {
			Required []string `json:"required"`
		}
		_ = json.Unmarshal(schemaBytes, &schema)

		args := map[string]any{}
		satisfiable := true
		for _, r := range schema.Required {
			v, found := fill[r]
			if !found || v == "" {
				satisfiable = false
				break
			}
			args[r] = v
		}
		// Skip team-scoped tools when no team was discovered, etc.
		if !satisfiable {
			skipped++
			t.Logf("SKIP  %-34s (needs %v)", tool.Name, schema.Required)
			continue
		}

		res := callTool(t, cs, tool.Name, args)
		ran++
		switch {
		case !res.IsError:
			ok++
			t.Logf("OK    %s", tool.Name)
		case tolerated(errText(res)):
			tol++
			t.Logf("TOL   %-34s %s", tool.Name, truncate(errText(res), 120))
		default:
			failures = append(failures, tool.Name+": "+truncate(errText(res), 200))
			t.Errorf("FAIL  %s -> %s", tool.Name, truncate(errText(res), 200))
		}
	}

	t.Logf("sweep: ran=%d ok=%d tolerated=%d skipped=%d failed=%d", ran, ok, tol, skipped, len(failures))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
