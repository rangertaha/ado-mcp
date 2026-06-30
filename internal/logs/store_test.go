// SPDX-License-Identifier: MIT

package logs

import (
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/gorm"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_AddGet(t *testing.T) {
	s := openTemp(t)
	e, err := s.Add(&Entry{Date: "2026-06-29", Summary: "Fixed the parser", Minutes: 90, Project: "Core"})
	if err != nil {
		t.Fatal(err)
	}
	if e.ID == 0 {
		t.Fatal("expected an assigned ID")
	}
	got, err := s.Get(e.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != "Fixed the parser" || got.Minutes != 90 {
		t.Errorf("got %+v", got)
	}
}

func TestStore_ListFilters(t *testing.T) {
	s := openTemp(t)
	_, _ = s.Add(&Entry{Date: "2026-06-28", Summary: "a"})
	_, _ = s.Add(&Entry{Date: "2026-06-29", Summary: "b"})
	logged, _ := s.Add(&Entry{Date: "2026-06-29", Summary: "c", HoursLogged: true})

	// Exact date.
	day, err := s.List(ListFilter{Date: "2026-06-29"})
	if err != nil {
		t.Fatal(err)
	}
	if len(day) != 2 {
		t.Errorf("date filter returned %d, want 2", len(day))
	}

	// Range.
	rng, _ := s.List(ListFilter{From: "2026-06-28", To: "2026-06-28"})
	if len(rng) != 1 || rng[0].Summary != "a" {
		t.Errorf("range filter = %+v", rng)
	}

	// Unlogged only (excludes the logged entry c).
	un, _ := s.List(ListFilter{Unlogged: true})
	for _, e := range un {
		if e.ID == logged.ID {
			t.Errorf("unlogged filter returned a logged entry: %+v", e)
		}
	}
	if len(un) != 2 {
		t.Errorf("unlogged returned %d, want 2", len(un))
	}
}

func TestStore_UpdateMarksLogged(t *testing.T) {
	s := openTemp(t)
	e, _ := s.Add(&Entry{Date: "2026-06-29", Summary: "work", Minutes: 30})

	updated, err := s.Update(e.ID, map[string]any{"hours_logged": true, "work_item_id": 1234})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.HoursLogged || updated.WorkItemID != 1234 {
		t.Errorf("update not applied: %+v", updated)
	}
	// Other fields preserved.
	if updated.Summary != "work" || updated.Minutes != 30 {
		t.Errorf("update clobbered fields: %+v", updated)
	}
}

func TestStore_Summarize(t *testing.T) {
	s := openTemp(t)
	_, _ = s.Add(&Entry{Date: "2026-06-28", Summary: "a", Minutes: 60, Project: "Core", TicketCreated: true})
	_, _ = s.Add(&Entry{Date: "2026-06-29", Summary: "b", Minutes: 30, Project: "Core", HoursLogged: true})
	_, _ = s.Add(&Entry{Date: "2026-06-29", Summary: "c", Minutes: 90, Project: "Web"})

	sum, err := s.Summarize(ListFilter{From: "2026-06-28", To: "2026-06-29"})
	if err != nil {
		t.Fatal(err)
	}
	if sum.Entries != 3 || sum.TotalMinutes != 180 || sum.TotalHours != 3 {
		t.Errorf("totals = %+v", sum)
	}
	if sum.TicketsCreated != 1 || sum.HoursLogged != 1 {
		t.Errorf("flags = tickets %d hours %d", sum.TicketsCreated, sum.HoursLogged)
	}
	if len(sum.ByDate) != 2 || sum.ByDate[0].Key != "2026-06-29" || sum.ByDate[0].Minutes != 120 {
		t.Errorf("byDate = %+v", sum.ByDate)
	}
	// Core (90) and Web (90) tie on minutes; Core sorts first alphabetically.
	if len(sum.ByProject) != 2 || sum.ByProject[0].Key != "Core" || sum.ByProject[0].Minutes != 90 {
		t.Errorf("byProject = %+v", sum.ByProject)
	}
}

func TestStore_Delete(t *testing.T) {
	s := openTemp(t)
	e, _ := s.Add(&Entry{Date: "2026-06-29", Summary: "x"})
	if err := s.Delete(e.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(e.ID); err == nil {
		t.Error("expected error getting a deleted entry")
	}
}

func TestStore_DeleteMissingReturnsNotFound(t *testing.T) {
	s := openTemp(t)
	// Deleting an ID that never existed must report not-found rather than
	// silently succeeding (GORM treats a zero-row delete as success).
	if err := s.Delete(999); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}
