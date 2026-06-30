// SPDX-License-Identifier: MIT

package sevenpace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(t *testing.T, h http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	c, err := NewClient(srv.URL, "tok")
	if err != nil {
		t.Fatal(err)
	}
	return c, srv
}

func TestListWorkLogs_UnwrapsODataValueAndSendsBearer(t *testing.T) {
	var auth string
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		if r.URL.Path != "/"+entityWorkLogsOnly {
			t.Errorf("path = %q, want /%s", r.URL.Path, entityWorkLogsOnly)
		}
		q := r.URL.Query()
		if q.Get("$filter") != "Timestamp ge 2026-01-01T00:00:00Z" || q.Get("$top") != "5" {
			t.Errorf("odata options = %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"value":[{"Id":"w1","Length":3600},{"Id":"w2","Length":1800}]}`))
	})
	defer srv.Close()

	logs, err := c.ListWorkLogs(context.Background(), "Timestamp ge 2026-01-01T00:00:00Z", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 || logs[0]["Id"] != "w1" {
		t.Errorf("logs = %+v", logs)
	}
	if auth != "Bearer tok" {
		t.Errorf("auth = %q, want Bearer tok", auth)
	}
}

func TestListWorkItems_UsesWorkItemsEntity(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+entityWorkItems {
			t.Errorf("path = %q, want /%s", r.URL.Path, entityWorkItems)
		}
		_, _ = w.Write([]byte(`{"value":[{"System_Id":42}]}`))
	})
	defer srv.Close()

	items, err := c.ListWorkItems(context.Background(), "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Errorf("items = %+v", items)
	}
}

func TestQueryRaw_ReturnsBodyVerbatim(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/$metadata" {
			t.Errorf("path = %q", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"$Version":"4.0"}`))
	})
	defer srv.Close()

	body, err := c.QueryRaw(context.Background(), "$metadata", nil)
	if err != nil {
		t.Fatal(err)
	}
	if body != `{"$Version":"4.0"}` {
		t.Errorf("body = %q", body)
	}
}
