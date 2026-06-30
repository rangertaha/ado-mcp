// SPDX-License-Identifier: MIT

package macros

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

// CompletePullRequest must GET the PR for its last merge commit, then PATCH it
// to completed using that commit.
func TestCompletePullRequest(t *testing.T) {
	var patchBody map[string]any
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"pullRequestId":7,"lastMergeSourceCommit":{"commitId":"deadbeef"}}`))
		case http.MethodPatch:
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &patchBody)
			_, _ = w.Write([]byte(`{"pullRequestId":7,"status":"completed"}`))
		}
	})
	defer srv.Close()

	out, err := svc.CompletePullRequest(context.Background(), "proj", "repo", 7, true, false, "squash")
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "completed" {
		t.Errorf("status = %q", out.Status)
	}
	if patchBody["status"] != "completed" {
		t.Errorf("patch status = %v", patchBody["status"])
	}
	lmsc := patchBody["lastMergeSourceCommit"].(map[string]any)
	if lmsc["commitId"] != "deadbeef" {
		t.Errorf("lastMergeSourceCommit = %v", lmsc)
	}
	opts := patchBody["completionOptions"].(map[string]any)
	if opts["deleteSourceBranch"] != true || opts["mergeStrategy"] != "squash" {
		t.Errorf("completionOptions = %v", opts)
	}
}

func TestCompletePullRequest_NotMergeable(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"pullRequestId":7}`)) // no lastMergeSourceCommit
	})
	defer srv.Close()

	if _, err := svc.CompletePullRequest(context.Background(), "proj", "repo", 7, false, false, ""); err == nil {
		t.Fatal("expected error when PR has no last merge source commit")
	}
}

// CreateBug posts a JSON-patch create, then a comment when one is supplied.
func TestCreateBug_WithComment(t *testing.T) {
	var sawCreate, sawComment bool
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/_apis/wit/workitems/$Bug"):
			sawCreate = true
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json-patch+json") {
				t.Errorf("create content-type = %q", r.Header.Get("Content-Type"))
			}
			_, _ = w.Write([]byte(`{"id":7}`))
		case strings.Contains(r.URL.Path, "/_apis/wit/workItems/7/comments"):
			sawComment = true
			_, _ = w.Write([]byte(`{"id":1,"text":"note"}`))
		default:
			t.Errorf("unexpected path %q", r.URL.Path)
		}
	})
	defer srv.Close()

	wi, err := svc.CreateBug(context.Background(), "proj", "Boom", "repro", "2 - High", "", "", "", "note")
	if err != nil {
		t.Fatal(err)
	}
	if wi.ID != 7 || !sawCreate || !sawComment {
		t.Errorf("id=%d create=%v comment=%v", wi.ID, sawCreate, sawComment)
	}
}

// PublishWikiPage on an existing page must send If-Match with the page ETag and
// report created=false.
func TestPublishWikiPage_UpdateUsesIfMatch(t *testing.T) {
	var putIfMatch string
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("ETag", `"v9"`)
			_, _ = w.Write([]byte(`{"path":"/P","content":"old"}`))
		case http.MethodPut:
			putIfMatch = r.Header.Get("If-Match")
			_, _ = w.Write([]byte(`{"path":"/P"}`))
		}
	})
	defer srv.Close()

	created, _, err := svc.PublishWikiPage(context.Background(), "proj", "wiki", "/P", "new")
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Error("expected created=false for an existing page")
	}
	if putIfMatch != `"v9"` {
		t.Errorf("If-Match = %q, want \"v9\"", putIfMatch)
	}
}

// PublishWikiPageWithImages uploads attachments and substitutes {{att:NAME}}.
func TestPublishWikiPageWithImages_SubstitutesPlaceholder(t *testing.T) {
	var publishedContent string
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/attachments"):
			_, _ = w.Write([]byte(`{"name":"diagram.png","path":"/.attachments/diagram.png"}`))
		case r.Method == http.MethodGet: // page existence probe
			http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
		case r.Method == http.MethodPut: // publish
			var body map[string]string
			b, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(b, &body)
			publishedContent = body["content"]
			_, _ = w.Write([]byte(`{"path":"/Design"}`))
		}
	})
	defer srv.Close()

	created, _, atts, err := svc.PublishWikiPageWithImages(
		context.Background(), "proj", "wiki", "/Design",
		"See ![diagram]({{att:diagram.png}})",
		[]WikiImage{{Name: "diagram.png", Content: "Ym9 G="}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Error("expected created=true for a new page")
	}
	if !strings.Contains(publishedContent, "/.attachments/diagram.png") {
		t.Errorf("placeholder not substituted: %q", publishedContent)
	}
	if len(atts) != 1 || atts[0].Path != "/.attachments/diagram.png" {
		t.Errorf("attachments = %+v", atts)
	}
}
