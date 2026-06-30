// SPDX-License-Identifier: MIT

package logs

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "logs"

// today returns the current date as YYYY-MM-DD.
func today() string { return time.Now().Format("2006-01-02") }

// RegisterTools adds the work-log journal toolset, backed by the SQLite store.
// It is a no-op when store is nil (database unavailable).
func RegisterTools(s *server.Server, store *Store) {
	if store == nil {
		return
	}
	s.NoteToolset(Name)
	h := &handlers{store: store}

	server.Register(s, server.ToolDef{
		Name:        "logs_add",
		Title:       "Add work-log entry",
		Description: "Record a daily work-log (journal) entry: what you worked on, optional minutes, project, linked work item, and 7pace activity. Date defaults to today.",
		Write:       true,
	}, h.add)

	server.Register(s, server.ToolDef{
		Name:        "logs_list",
		Title:       "List work-log entries",
		Description: "List work-log entries, newest first. Filter by an exact date, a date range, or only entries whose hours have not yet been logged to 7pace.",
	}, h.list)

	server.Register(s, server.ToolDef{
		Name:        "logs_get",
		Title:       "Get work-log entry",
		Description: "Get a single work-log entry by ID.",
	}, h.get)

	server.Register(s, server.ToolDef{
		Name:        "logs_update",
		Title:       "Update work-log entry",
		Description: "Update fields of a work-log entry, including marking that a ticket was created or hours were logged. Omitted fields are left unchanged.",
		Write:       true,
		Idempotent:  true,
	}, h.update)

	server.Register(s, server.ToolDef{
		Name:        "logs_delete",
		Title:       "Delete work-log entry",
		Description: "Delete a work-log entry by ID.",
		Write:       true,
		Destructive: true,
		Idempotent:  true,
	}, h.delete)

	server.Register(s, server.ToolDef{
		Name:        "logs_summary",
		Title:       "Summarize work-log entries",
		Description: "Aggregate work-log entries over a date (or date range): total entries and minutes/hours, tickets created, hours logged, plus per-day and per-project breakdowns.",
	}, h.summary)
}

type handlers struct{ store *Store }

// --- Inputs ---

// AddInput records a new work-log entry.
type AddInput struct {
	Summary    string `json:"summary" jsonschema:"what you worked on"`
	Date       string `json:"date,omitempty" jsonschema:"work date YYYY-MM-DD (defaults to today)"`
	Minutes    int    `json:"minutes,omitempty" jsonschema:"time spent, in minutes"`
	Project    string `json:"project,omitempty" jsonschema:"Azure DevOps project"`
	WorkItemID int    `json:"workItemId,omitempty" jsonschema:"linked Azure DevOps work item ID"`
	Activity   string `json:"activity,omitempty" jsonschema:"7pace activity type"`
	Tags       string `json:"tags,omitempty" jsonschema:"comma-separated tags"`
}

// ListInput filters work-log entries.
type ListInput struct {
	Date     string `json:"date,omitempty" jsonschema:"exact work date YYYY-MM-DD"`
	From     string `json:"from,omitempty" jsonschema:"inclusive start date YYYY-MM-DD"`
	To       string `json:"to,omitempty" jsonschema:"inclusive end date YYYY-MM-DD"`
	Unlogged bool   `json:"unlogged,omitempty" jsonschema:"only entries whose hours are not yet logged to 7pace"`
	Limit    int    `json:"limit,omitempty" jsonschema:"maximum entries (default 200)"`
}

// IDInput identifies an entry.
type IDInput struct {
	ID uint `json:"id" jsonschema:"work-log entry ID"`
}

// SummaryInput scopes a work-log summary.
type SummaryInput struct {
	Date string `json:"date,omitempty" jsonschema:"exact work date YYYY-MM-DD"`
	From string `json:"from,omitempty" jsonschema:"inclusive start date YYYY-MM-DD"`
	To   string `json:"to,omitempty" jsonschema:"inclusive end date YYYY-MM-DD"`
}

// UpdateInput changes fields of an entry. Pointer fields are applied only when
// provided, so callers can update a single field without clearing others.
type UpdateInput struct {
	ID            uint    `json:"id" jsonschema:"work-log entry ID"`
	Summary       *string `json:"summary,omitempty" jsonschema:"new summary"`
	Date          *string `json:"date,omitempty" jsonschema:"new work date YYYY-MM-DD"`
	Minutes       *int    `json:"minutes,omitempty" jsonschema:"new minutes"`
	Project       *string `json:"project,omitempty" jsonschema:"new project"`
	WorkItemID    *int    `json:"workItemId,omitempty" jsonschema:"linked work item ID"`
	Activity      *string `json:"activity,omitempty" jsonschema:"7pace activity type"`
	Tags          *string `json:"tags,omitempty" jsonschema:"comma-separated tags"`
	TicketCreated *bool   `json:"ticketCreated,omitempty" jsonschema:"mark that a work item was created"`
	HoursLogged   *bool   `json:"hoursLogged,omitempty" jsonschema:"mark that hours were logged to 7pace"`
}

// --- Handlers ---

func (h *handlers) add(_ context.Context, _ *mcp.CallToolRequest, in AddInput) (*mcp.CallToolResult, *Entry, error) {
	date := in.Date
	if date == "" {
		date = today()
	}
	out, err := h.store.Add(&Entry{
		Date:       date,
		Summary:    in.Summary,
		Minutes:    in.Minutes,
		Project:    in.Project,
		WorkItemID: in.WorkItemID,
		Activity:   in.Activity,
		Tags:       in.Tags,
	})
	return nil, out, err
}

func (h *handlers) list(_ context.Context, _ *mcp.CallToolRequest, in ListInput) (*mcp.CallToolResult, server.ListResult[Entry], error) {
	out, err := h.store.List(ListFilter{Date: in.Date, From: in.From, To: in.To, Unlogged: in.Unlogged, Limit: in.Limit})
	return nil, server.List(out), err
}

func (h *handlers) get(_ context.Context, _ *mcp.CallToolRequest, in IDInput) (*mcp.CallToolResult, *Entry, error) {
	out, err := h.store.Get(in.ID)
	return nil, out, err
}

func (h *handlers) update(_ context.Context, _ *mcp.CallToolRequest, in UpdateInput) (*mcp.CallToolResult, *Entry, error) {
	changes := map[string]any{}
	if in.Summary != nil {
		changes["summary"] = *in.Summary
	}
	if in.Date != nil {
		changes["date"] = *in.Date
	}
	if in.Minutes != nil {
		changes["minutes"] = *in.Minutes
	}
	if in.Project != nil {
		changes["project"] = *in.Project
	}
	if in.WorkItemID != nil {
		changes["work_item_id"] = *in.WorkItemID
	}
	if in.Activity != nil {
		changes["activity"] = *in.Activity
	}
	if in.Tags != nil {
		changes["tags"] = *in.Tags
	}
	if in.TicketCreated != nil {
		changes["ticket_created"] = *in.TicketCreated
	}
	if in.HoursLogged != nil {
		changes["hours_logged"] = *in.HoursLogged
	}
	out, err := h.store.Update(in.ID, changes)
	return nil, out, err
}

func (h *handlers) delete(_ context.Context, _ *mcp.CallToolRequest, in IDInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := h.store.Delete(in.ID); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

func (h *handlers) summary(_ context.Context, _ *mcp.CallToolRequest, in SummaryInput) (*mcp.CallToolResult, *Summary, error) {
	out, err := h.store.Summarize(ListFilter{Date: in.Date, From: in.From, To: in.To})
	return nil, out, err
}
