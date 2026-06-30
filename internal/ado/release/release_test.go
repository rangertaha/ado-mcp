// SPDX-License-Identifier: MIT

package release

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// release uses the VSRM host client.
func newTestService(t *testing.T, h http.HandlerFunc) (*service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c, err := client.New(srv.URL, client.NewPATAuthorizer("t"), client.WithAPIVersion("7.1"))
	if err != nil {
		t.Fatal(err)
	}
	return &service{c: &ado.Clients{VSRM: c}}, srv
}

func TestListReleases_ForwardsFilters(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proj/_apis/release/releases" {
			t.Errorf("path = %q", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("definitionId") != "9" || q.Get("$top") != "3" {
			t.Errorf("query = %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"count":1,"value":[{"id":11,"name":"Release-11","status":"active"}]}`))
	})
	defer srv.Close()

	out, err := svc.ListReleases(context.Background(), "proj", 9, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].ID != 11 || out[0].Status != "active" {
		t.Errorf("got %+v", out)
	}
}

func TestCreateRelease_PostsDefinitionAndDescription(t *testing.T) {
	var body map[string]any
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/proj/_apis/release/releases" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		_, _ = w.Write([]byte(`{"id":12,"name":"Release-12"}`))
	})
	defer srv.Close()

	out, err := svc.CreateRelease(context.Background(), "proj", 9, "ship it")
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != 12 {
		t.Errorf("id = %d", out.ID)
	}
	// definitionId is JSON-decoded as a float64.
	if body["definitionId"].(float64) != 9 || body["description"] != "ship it" {
		t.Errorf("body = %v", body)
	}
}

func TestUpdateApproval_PatchesStatus(t *testing.T) {
	var body map[string]any
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/proj/_apis/release/approvals/55" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		_, _ = w.Write([]byte(`{"id":55,"status":"approved"}`))
	})
	defer srv.Close()

	out, err := svc.UpdateApproval(context.Background(), "proj", 55, "approved", "ok")
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "approved" {
		t.Errorf("status = %q", out.Status)
	}
	if body["status"] != "approved" || body["comments"] != "ok" {
		t.Errorf("body = %v", body)
	}
}
