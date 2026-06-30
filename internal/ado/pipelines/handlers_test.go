// SPDX-License-Identifier: MIT

package pipelines

import (
	"context"
	"net/http"
	"testing"
)

// These tests exercise the MCP tool handler layer: input struct -> service ->
// structured output (including ListResult wrapping), with a nil request.

func TestHandler_listPipelines_WrapsList(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"count":2,"value":[{"id":1,"name":"a"},{"id":2,"name":"b"}]}`))
	})
	defer srv.Close()

	_, out, err := svc.listPipelines(context.Background(), nil, ListPipelinesInput{Project: "p", Top: 10})
	if err != nil {
		t.Fatal(err)
	}
	if out.Count != 2 || len(out.Items) != 2 || out.Items[0].Name != "a" {
		t.Errorf("ListResult = %+v", out)
	}
}

func TestHandler_getBuild_ReturnsPointer(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":5,"result":"succeeded"}`))
	})
	defer srv.Close()

	_, out, err := svc.getBuild(context.Background(), nil, BuildInput{Project: "p", BuildID: 5})
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || out.ID != 5 || out.Result != "succeeded" {
		t.Errorf("build = %+v", out)
	}
}

func TestHandler_getBuildLog_WrapsTextInObject(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("log body"))
	})
	defer srv.Close()

	_, out, err := svc.getBuildLog(context.Background(), nil, BuildLogInput{Project: "p", BuildID: 5, LogID: 1})
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || out.Log != "log body" {
		t.Errorf("log output = %+v", out)
	}
}

func TestHandler_runPipeline_PropagatesError(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"pipeline not found"}`))
	})
	defer srv.Close()

	_, _, err := svc.runPipeline(context.Background(), nil, RunPipelineInput{Project: "p", PipelineID: 99, Branch: "main"})
	if err == nil {
		t.Fatal("expected error to propagate from handler")
	}
}
