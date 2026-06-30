// SPDX-License-Identifier: MIT

package repostats

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

func TestAggregator_TalliesAndSorts(t *testing.T) {
	agg := newAggregator()
	agg.add(commit{Author: signature{Name: "Ada", Email: "ada@x"}, ChangeCounts: changeCounts{Add: 2, Edit: 1}})
	agg.add(commit{Author: signature{Name: "Ada", Email: "ada@x", Date: "2026-01-05"}, ChangeCounts: changeCounts{Edit: 3}})
	agg.add(commit{Author: signature{Name: "Ada", Email: "ada@x", Date: "2026-01-02"}})
	agg.add(commit{Author: signature{Name: "Bo", Email: "bo@x", Date: "2026-02-01"}})

	st := agg.result("repo:Foo", 1, false)

	if st.CommitsScanned != 4 {
		t.Errorf("CommitsScanned = %d, want 4", st.CommitsScanned)
	}
	if len(st.Contributors) != 2 {
		t.Fatalf("contributors = %d, want 2", len(st.Contributors))
	}
	// Ada has 3 commits and must sort first.
	ada := st.Contributors[0]
	if ada.Email != "ada@x" || ada.Commits != 3 {
		t.Errorf("top contributor = %+v, want Ada with 3 commits", ada)
	}
	if ada.Adds != 2 || ada.Edits != 4 {
		t.Errorf("Ada change counts = adds %d edits %d, want 2/4", ada.Adds, ada.Edits)
	}
	if ada.FirstCommit != "2026-01-02" || ada.LastCommit != "2026-01-05" {
		t.Errorf("Ada dates = %s..%s, want 2026-01-02..2026-01-05", ada.FirstCommit, ada.LastCommit)
	}
	if st.Contributors[1].Email != "bo@x" {
		t.Errorf("second contributor = %s, want bo@x", st.Contributors[1].Email)
	}
}

func TestAggregator_FallsBackToNameWhenNoEmail(t *testing.T) {
	agg := newAggregator()
	agg.add(commit{Author: signature{Name: "NoEmail"}})
	agg.add(commit{Author: signature{Name: "NoEmail"}})
	st := agg.result("x", 1, false)
	if len(st.Contributors) != 1 || st.Contributors[0].Commits != 2 {
		t.Errorf("expected a single contributor with 2 commits, got %+v", st.Contributors)
	}
}

func TestPRAggregator_CountsByStatus(t *testing.T) {
	agg := newPRAggregator()
	mk := func(name, status string) prWire {
		var pr prWire
		pr.Status = status
		pr.CreatedBy.DisplayName = name
		pr.CreatedBy.UniqueName = name + "@x"
		return pr
	}
	agg.add(mk("Ada", "active"))
	agg.add(mk("Ada", "completed"))
	agg.add(mk("Ada", "completed"))
	agg.add(mk("Bo", "abandoned"))

	st := agg.result("project:P", 2, false)

	if st.PRsScanned != 4 {
		t.Errorf("PRsScanned = %d, want 4", st.PRsScanned)
	}
	if st.ByStatus["completed"] != 2 || st.ByStatus["active"] != 1 || st.ByStatus["abandoned"] != 1 {
		t.Errorf("ByStatus = %v", st.ByStatus)
	}
	top := st.Authors[0]
	if top.UniqueName != "Ada@x" || top.Total != 3 || top.Completed != 2 || top.Active != 1 {
		t.Errorf("top author = %+v, want Ada 3/active1/completed2", top)
	}
}

func TestFieldString_ResolvesIdentity(t *testing.T) {
	fields := map[string]any{
		"System.State":      "Active",
		"System.AssignedTo": map[string]any{"displayName": "Ada Lovelace", "uniqueName": "ada@x"},
		"System.Missing":    nil,
	}
	if got := fieldString(fields, "System.State"); got != "Active" {
		t.Errorf("state = %q", got)
	}
	if got := fieldString(fields, "System.AssignedTo"); got != "Ada Lovelace" {
		t.Errorf("assignee = %q, want display name", got)
	}
	if got := fieldString(fields, "System.Missing"); got != "" {
		t.Errorf("missing = %q, want empty", got)
	}
	if got := fieldString(fields, "System.Absent"); got != "" {
		t.Errorf("absent = %q, want empty", got)
	}
}

// TestBuildStats_HonorsMaxBuildsCap verifies the scan clamps its page size to
// the remaining budget instead of pulling a full page and over-tallying.
func TestBuildStats_HonorsMaxBuildsCap(t *testing.T) {
	var tops []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tops = append(tops, r.URL.Query().Get("$top"))
		// Return exactly as many builds as requested via $top.
		n, _ := strconv.Atoi(r.URL.Query().Get("$top"))
		builds := make([]map[string]any, n)
		for i := range builds {
			builds[i] = map[string]any{"id": i, "result": "succeeded",
				"definition": map[string]any{"id": 1, "name": "CI"}}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"value": builds})
	}))
	defer srv.Close()

	c, err := client.New(srv.URL, client.NewPATAuthorizer("t"), client.WithAPIVersion("7.1"))
	if err != nil {
		t.Fatal(err)
	}
	svc := &service{c: &ado.Clients{Org: c}}

	st, err := svc.BuildStatsForProject(context.Background(), "proj", 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if st.BuildsScanned != 10 {
		t.Errorf("BuildsScanned = %d, want 10 (cap must be honored)", st.BuildsScanned)
	}
	if len(tops) != 1 || tops[0] != "10" {
		t.Errorf("$top sequence = %v, want a single clamped page of 10", tops)
	}
}
