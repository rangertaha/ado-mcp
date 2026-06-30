// SPDX-License-Identifier: MIT

package release

import (
	"context"
	"net/http"
	"testing"
)

// Handler-layer tests: input struct -> service (VSRM host) -> structured output.

func TestHandler_listReleases_WrapsList(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"count":1,"value":[{"id":11,"name":"R-11","status":"active"}]}`))
	})
	defer srv.Close()

	_, out, err := svc.listReleases(context.Background(), nil, ListReleasesInput{Project: "p", DefinitionID: 9, Top: 3})
	if err != nil {
		t.Fatal(err)
	}
	if out.Count != 1 || len(out.Items) != 1 || out.Items[0].ID != 11 {
		t.Errorf("ListResult = %+v", out)
	}
}

func TestHandler_createRelease_ReturnsPointer(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":12,"name":"R-12"}`))
	})
	defer srv.Close()

	_, out, err := svc.createRelease(context.Background(), nil, CreateReleaseInput{Project: "p", DefinitionID: 9, Description: "go"})
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || out.ID != 12 {
		t.Errorf("release = %+v", out)
	}
}

func TestHandler_updateApproval_ReturnsPointer(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"id":55,"status":"approved"}`))
	})
	defer srv.Close()

	_, out, err := svc.updateApproval(context.Background(), nil, UpdateApprovalInput{Project: "p", ApprovalID: 55, Status: "approved", Comments: "ok"})
	if err != nil {
		t.Fatal(err)
	}
	if out == nil || out.Status != "approved" {
		t.Errorf("approval = %+v", out)
	}
}

func TestHandler_listApprovals_PropagatesError(t *testing.T) {
	svc, srv := newTestService(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"message":"access denied"}`))
	})
	defer srv.Close()

	_, _, err := svc.listApprovals(context.Background(), nil, ListApprovalsInput{Project: "p", StatusFilter: "pending"})
	if err == nil {
		t.Fatal("expected error to propagate from handler")
	}
}
