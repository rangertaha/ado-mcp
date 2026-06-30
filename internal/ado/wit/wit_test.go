// SPDX-License-Identifier: MIT

package wit

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

func TestGetWorkItem(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proj/_apis/wit/workitems/42" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("$expand") != "all" {
			t.Errorf("expand = %q", r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"id":42,"rev":3,"fields":{"System.Title":"Boom"}}`))
	})
	defer srv.Close()

	wi, err := svc.GetWorkItem(context.Background(), "proj", 42, "all")
	if err != nil {
		t.Fatal(err)
	}
	if wi.ID != 42 || wi.Fields["System.Title"] != "Boom" {
		t.Errorf("got %+v", wi)
	}
}

func TestQuery_PostsWIQL(t *testing.T) {
	var body map[string]string
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/proj/_apis/wit/wiql" {
			t.Errorf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"queryType":"flat","workItems":[{"id":1},{"id":2}]}`))
	})
	defer srv.Close()

	res, err := svc.Query(context.Background(), "proj", "SELECT [System.Id] FROM WorkItems", 50)
	if err != nil {
		t.Fatal(err)
	}
	if body["query"] != "SELECT [System.Id] FROM WorkItems" {
		t.Errorf("query body = %v", body)
	}
	if len(res.WorkItems) != 2 || res.WorkItems[1].ID != 2 {
		t.Errorf("result = %+v", res)
	}
}

func TestCreateWorkItem_SendsJSONPatch(t *testing.T) {
	var ct string
	var ops []map[string]any
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/proj/_apis/wit/workitems/$Bug" {
			t.Errorf("path = %q", r.URL.Path)
		}
		ct = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &ops)
		_, _ = w.Write([]byte(`{"id":99}`))
	})
	defer srv.Close()

	wi, err := svc.CreateWorkItem(context.Background(), "proj", "Bug", map[string]any{"System.Title": "X"})
	if err != nil {
		t.Fatal(err)
	}
	if wi.ID != 99 {
		t.Errorf("id = %d", wi.ID)
	}
	if !strings.HasPrefix(ct, "application/json-patch+json") {
		t.Errorf("content-type = %q, want json-patch", ct)
	}
	if len(ops) != 1 || ops[0]["op"] != "add" || ops[0]["path"] != "/fields/System.Title" {
		t.Errorf("patch ops = %+v", ops)
	}
}
