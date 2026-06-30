// SPDX-License-Identifier: MIT

package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// newTestService builds a core service whose Org client points at the given URL.
func newTestService(t *testing.T, baseURL string) *service {
	t.Helper()
	c, err := client.New(baseURL, client.NewPATAuthorizer("t"), client.WithAPIVersion("7.1"))
	if err != nil {
		t.Fatal(err)
	}
	return &service{c: &ado.Clients{Org: c}}
}

func TestListProjects(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_apis/projects" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.URL.Query().Get("$top") != "2" {
			t.Errorf("$top not forwarded: %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"count":2,"value":[{"id":"1","name":"Alpha"},{"id":"2","name":"Beta"}]}`))
	}))
	defer srv.Close()

	svc := newTestService(t, srv.URL)
	got, err := svc.ListProjects(context.Background(), 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "Alpha" || got[1].ID != "2" {
		t.Errorf("ListProjects = %+v", got)
	}
}

func TestGetProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_apis/projects/Alpha" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"1","name":"Alpha","state":"wellFormed"}`))
	}))
	defer srv.Close()

	svc := newTestService(t, srv.URL)
	got, err := svc.GetProject(context.Background(), "Alpha")
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Alpha" || got.State != "wellFormed" {
		t.Errorf("GetProject = %+v", got)
	}
}

func TestListTeamMembers_UnwrapsIdentity(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"count":1,"value":[{"identity":{"id":"u1","displayName":"Ada"}}]}`))
	}))
	defer srv.Close()

	svc := newTestService(t, srv.URL)
	got, err := svc.ListTeamMembers(context.Background(), "Alpha", "TeamA")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].DisplayName != "Ada" {
		t.Errorf("ListTeamMembers = %+v", got)
	}
}
