// SPDX-License-Identifier: MIT

package boards

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

func newTestService(t *testing.T, h http.HandlerFunc) (*service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c, err := client.New(srv.URL, client.NewPATAuthorizer("t"), client.WithAPIVersion("7.1"))
	if err != nil {
		t.Fatal(err)
	}
	return &service{c: &ado.Clients{Org: c}}, srv
}

// Boards endpoints are team-scoped: /{project}/{team}/_apis/work/...
func TestListBacklogs_TeamScopedPath(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proj/teamA/_apis/work/backlogs" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"count":2,"value":[{"id":"Epics","name":"Epics","rank":1},{"id":"Stories","name":"Stories","rank":2}]}`))
	})
	defer srv.Close()

	out, err := svc.ListBacklogs(context.Background(), "proj", "teamA")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].ID != "Epics" || out[1].Rank != 2 {
		t.Errorf("got %+v", out)
	}
}

func TestGetBoardColumns(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proj/teamA/_apis/work/boards/Stories/columns" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"count":1,"value":[{"id":"c1","name":"Doing","columnType":"inProgress","itemLimit":5}]}`))
	})
	defer srv.Close()

	out, err := svc.GetBoardColumns(context.Background(), "proj", "teamA", "Stories")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Name != "Doing" || out[0].ItemLimit != 5 {
		t.Errorf("got %+v", out)
	}
}
