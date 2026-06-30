// SPDX-License-Identifier: MIT

package pipelines

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

func newTestService(t *testing.T, h http.HandlerFunc) (*service, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c, err := client.New(srv.URL, client.NewPATAuthorizer("t"), client.WithAPIVersion("7.1"))
	if err != nil {
		t.Fatal(err)
	}
	return &service{c: &ado.Clients{Org: c}}, srv
}

func TestListPipelines(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proj/_apis/pipelines" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("$top") != "5" {
			t.Errorf("$top = %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"count":1,"value":[{"id":3,"name":"CI"}]}`))
	})
	defer srv.Close()

	out, err := svc.ListPipelines(context.Background(), "proj", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].ID != 3 || out[0].Name != "CI" {
		t.Errorf("got %+v", out)
	}
}

// GetBuildLog returns a plain-text (non-JSON) body via the raw-body path.
func TestGetBuildLog_RawText(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proj/_apis/build/builds/5/logs/2" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/plain")
		_, _ = io.WriteString(w, "line1\nline2\n")
	})
	defer srv.Close()

	log, err := svc.GetBuildLog(context.Background(), "proj", 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if log != "line1\nline2\n" {
		t.Errorf("log = %q", log)
	}
}

func TestCancelBuild_PatchesStatus(t *testing.T) {
	var body map[string]any
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s", r.Method)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &body)
		_, _ = w.Write([]byte(`{"id":5,"status":"cancelling"}`))
	})
	defer srv.Close()

	out, err := svc.CancelBuild(context.Background(), "proj", 5)
	if err != nil {
		t.Fatal(err)
	}
	if body["status"] != "cancelling" {
		t.Errorf("body = %v", body)
	}
	if out.ID != 5 {
		t.Errorf("out = %+v", out)
	}
}
