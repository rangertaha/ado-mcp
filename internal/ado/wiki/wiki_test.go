// SPDX-License-Identifier: MIT

package wiki

import (
	"context"
	"encoding/json"
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

// On a missing page, CreateOrUpdatePage must create it WITHOUT an If-Match header.
func TestCreateOrUpdatePage_CreatesWithoutIfMatch(t *testing.T) {
	var putIfMatch string
	var putSeen bool
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			http.Error(w, `{"message":"page not found"}`, http.StatusNotFound)
		case http.MethodPut:
			putSeen = true
			putIfMatch = r.Header.Get("If-Match")
			_, _ = w.Write([]byte(`{"path":"/New","content":"hello"}`))
		}
	})
	defer srv.Close()

	p, err := svc.CreateOrUpdatePage(context.Background(), "proj", "wiki", "/New", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !putSeen {
		t.Fatal("expected a PUT to create the page")
	}
	if putIfMatch != "" {
		t.Errorf("create must not send If-Match, got %q", putIfMatch)
	}
	if p.Path != "/New" {
		t.Errorf("unexpected page: %+v", p)
	}
}

// On an existing page, CreateOrUpdatePage must send If-Match with the page ETag.
func TestCreateOrUpdatePage_UpdatesWithIfMatch(t *testing.T) {
	const etag = `"abc123"`
	var putIfMatch string
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("ETag", etag)
			_, _ = w.Write([]byte(`{"path":"/Existing","content":"old"}`))
		case http.MethodPut:
			putIfMatch = r.Header.Get("If-Match")
			_, _ = w.Write([]byte(`{"path":"/Existing","content":"new"}`))
		}
	})
	defer srv.Close()

	if _, err := svc.CreateOrUpdatePage(context.Background(), "proj", "wiki", "/Existing", "new"); err != nil {
		t.Fatal(err)
	}
	if putIfMatch != etag {
		t.Errorf("update must send If-Match=%s, got %q", etag, putIfMatch)
	}
}

// AppendToPage must concatenate existing content with the new content.
func TestAppendToPage_Concatenates(t *testing.T) {
	var putBody struct {
		Content string `json:"content"`
	}
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("ETag", `"v1"`)
			_, _ = w.Write([]byte(`{"path":"/Log","content":"old"}`))
		case http.MethodPut:
			_ = json.NewDecoder(r.Body).Decode(&putBody)
			_, _ = w.Write([]byte(`{"path":"/Log","content":"old\n\nnew"}`))
		}
	})
	defer srv.Close()

	if _, err := svc.AppendToPage(context.Background(), "proj", "wiki", "/Log", "new", ""); err != nil {
		t.Fatal(err)
	}
	if putBody.Content != "old\n\nnew" {
		t.Errorf("append produced %q, want %q", putBody.Content, "old\n\nnew")
	}
}
