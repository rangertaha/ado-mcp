// SPDX-License-Identifier: MIT

// Package logs implements a local daily work-log journal, persisted to SQLite
// (via GORM) under the user's application data directory.
//
// Entries record what was worked on (date, summary, minutes, optional project
// and work-item link). Claude reads them through the "logs" MCP toolset and
// turns them into Azure DevOps work items (tickets) and 7pace time entries,
// marking each entry as it is processed so work is not duplicated.
//
// The store is intentionally separate from server diagnostics: the server logs
// operational messages to stderr (stdout is reserved for the MCP protocol),
// while this package owns the user's persistent work journal.
package logs
