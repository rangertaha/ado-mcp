// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestDo_GETDecodesAndAddsAPIVersion(t *testing.T) {
	var gotAuth, gotAccept, gotAPIVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotAPIVersion = r.URL.Query().Get("api-version")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":2,"value":[{"name":"a"},{"name":"b"}]}`))
	}))
	defer srv.Close()

	c, err := New(srv.URL, NewPATAuthorizer("tok"), WithAPIVersion("7.1"))
	if err != nil {
		t.Fatal(err)
	}

	var out List[struct {
		Name string `json:"name"`
	}]
	if err := c.GetJSON(context.Background(), "/_apis/projects", nil, &out); err != nil {
		t.Fatalf("GetJSON: %v", err)
	}

	if out.Count != 2 || len(out.Value) != 2 || out.Value[0].Name != "a" {
		t.Errorf("unexpected decode: %+v", out)
	}
	wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(":tok"))
	if gotAuth != wantAuth {
		t.Errorf("auth header = %q, want %q", gotAuth, wantAuth)
	}
	if gotAccept != "application/json" {
		t.Errorf("accept header = %q", gotAccept)
	}
	if gotAPIVersion != "7.1" {
		t.Errorf("api-version = %q, want 7.1", gotAPIVersion)
	}
}

func TestDo_APIVersionOverride(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("api-version")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, _ := New(srv.URL, NewPATAuthorizer("t"), WithAPIVersion("7.1"))
	_, err := c.Do(context.Background(), Request{Method: http.MethodGet, Path: "/x", APIVersion: "7.1-preview.3", Out: &map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if got != "7.1-preview.3" {
		t.Errorf("api-version = %q, want override", got)
	}
}

func TestDo_ErrorEnvelopeIsParsed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"TF401232: work item does not exist","typeKey":"WorkItemDoesNotExist"}`))
	}))
	defer srv.Close()

	c, _ := New(srv.URL, NewBearerAuthorizer("t"))
	err := c.GetJSON(context.Background(), "/_apis/wit/workitems/9999", nil, &struct{}{})
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %v", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d", apiErr.StatusCode)
	}
	if apiErr.Message != "TF401232: work item does not exist" {
		t.Errorf("message = %q", apiErr.Message)
	}
	if apiErr.TypeKey != "WorkItemDoesNotExist" {
		t.Errorf("typeKey = %q", apiErr.TypeKey)
	}
}

func TestDo_PostSendsJSONBody(t *testing.T) {
	var gotBody, gotCT string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(b)
		gotBody = string(b)
		gotCT = r.Header.Get("Content-Type")
		_, _ = w.Write([]byte(`{"id":"new"}`))
	}))
	defer srv.Close()

	c, _ := New(srv.URL, NewBearerAuthorizer("t"))
	var out struct {
		ID string `json:"id"`
	}
	in := map[string]string{"name": "x"}
	if err := c.PostJSON(context.Background(), "/things", nil, in, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "new" {
		t.Errorf("id = %q", out.ID)
	}
	if gotCT != "application/json" {
		t.Errorf("content-type = %q", gotCT)
	}
	if gotBody != `{"name":"x"}` {
		t.Errorf("body = %q", gotBody)
	}
}

func TestContinuationTokenSurfaced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-ms-continuationtoken", "next-page")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, _ := New(srv.URL, NewBearerAuthorizer("t"))
	resp, err := c.Do(context.Background(), Request{Method: http.MethodGet, Path: "/", Query: url.Values{}, Out: &map[string]any{}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ContinuationToken != "next-page" {
		t.Errorf("continuation token = %q", resp.ContinuationToken)
	}
}

func TestDo_NegotiatesAPIVersionOnMismatch(t *testing.T) {
	var versions []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.URL.Query().Get("api-version")
		versions = append(versions, v)
		if v == "7.1" {
			w.Header().Set("Api-Supported-Versions", "5.0,6.0,7.0,7.1-preview.1")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"The requested REST API version of 7.1 is not supported."}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c, _ := New(srv.URL, NewPATAuthorizer("t"), WithAPIVersion("7.1"))
	var out map[string]any
	if err := c.GetJSON(context.Background(), "/_apis/audit/auditlog", nil, &out); err != nil {
		t.Fatalf("expected success after negotiation, got %v", err)
	}
	if len(versions) != 2 || versions[0] != "7.1" || versions[1] != "7.1-preview.1" {
		t.Errorf("expected retry with 7.1-preview.1, got sequence %v", versions)
	}
	if out["ok"] != true {
		t.Errorf("unexpected body: %v", out)
	}
}

func TestDo_DoesNotRetryNonVersion400(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Api-Supported-Versions", "7.1,7.1-preview.1")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"TF400898: invalid argument"}`))
	}))
	defer srv.Close()

	c, _ := New(srv.URL, NewPATAuthorizer("t"), WithAPIVersion("7.1"))
	err := c.GetJSON(context.Background(), "/x", nil, &map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("non-version 400 must not retry; got %d calls", calls)
	}
}
