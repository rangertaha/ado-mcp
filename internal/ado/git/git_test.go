// SPDX-License-Identifier: MIT

package git

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

func TestNormalizeRef(t *testing.T) {
	cases := map[string]string{
		"main":            "refs/heads/main",
		"feature/x":       "refs/heads/feature/x",
		"refs/heads/main": "refs/heads/main",
		"refs/tags/v1":    "refs/tags/v1",
		"":                "",
	}
	for in, want := range cases {
		if got := normalizeRef(in); got != want {
			t.Errorf("normalizeRef(%q) = %q, want %q", in, got, want)
		}
	}
}

// PushFile must first resolve the branch tip (oldObjectId) then post a push with
// that oldObjectId and the file change.
func TestPushFile_ResolvesBranchAndPushes(t *testing.T) {
	var pushBody map[string]any
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/proj/_apis/git/repositories/repo/refs":
			if r.URL.Query().Get("filter") != "heads/main" {
				t.Errorf("ref filter = %q", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"count":1,"value":[{"name":"refs/heads/main","objectId":"abc123"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/proj/_apis/git/repositories/repo/pushes":
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &pushBody)
			_, _ = w.Write([]byte(`{"pushId":7}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	})
	defer srv.Close()

	push, err := svc.PushFile(context.Background(), "proj", "repo", "main", "/a.txt", "hello", "msg", "add")
	if err != nil {
		t.Fatal(err)
	}
	if push.PushID != 7 {
		t.Errorf("pushId = %d", push.PushID)
	}
	refUpdates := pushBody["refUpdates"].([]any)
	ru := refUpdates[0].(map[string]any)
	if ru["name"] != "refs/heads/main" || ru["oldObjectId"] != "abc123" {
		t.Errorf("refUpdate = %v, want oldObjectId abc123", ru)
	}
}

// When the branch does not exist, PushFile uses the zero SHA to create it.
func TestPushFile_NewBranchUsesZeroSHA(t *testing.T) {
	var pushBody map[string]any
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"count":0,"value":[]}`))
		case http.MethodPost:
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &pushBody)
			_, _ = w.Write([]byte(`{"pushId":1}`))
		}
	})
	defer srv.Close()

	if _, err := svc.PushFile(context.Background(), "proj", "repo", "new", "/a", "x", "m", ""); err != nil {
		t.Fatal(err)
	}
	ru := pushBody["refUpdates"].([]any)[0].(map[string]any)
	if ru["oldObjectId"] != "0000000000000000000000000000000000000000" {
		t.Errorf("oldObjectId = %v, want zero SHA", ru["oldObjectId"])
	}
}
