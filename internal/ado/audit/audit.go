// SPDX-License-Identifier: MIT

// Package audit exposes the Azure DevOps Audit service: querying the
// organization-level audit log of decorated audit log entries.
package audit

import (
	"context"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "audit"

// service wraps the Azure DevOps clients for Audit operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// AuditLogEntry is a single decorated audit log entry.
type AuditLogEntry struct {
	ID               string `json:"id"`
	ActionID         string `json:"actionId,omitempty"`
	ActorDisplayName string `json:"actorDisplayName,omitempty"`
	ActorUPN         string `json:"actorUPN,omitempty"`
	Timestamp        string `json:"timestamp,omitempty"`
	ScopeType        string `json:"scopeType,omitempty"`
	ProjectName      string `json:"projectName,omitempty"`
	Details          string `json:"details,omitempty"`
	Area             string `json:"area,omitempty"`
	Category         string `json:"category,omitempty"`
}

// AuditLogResult is the response envelope returned by the audit log query.
// Only a single page of entries is returned; HasMore indicates that more
// pages exist and ContinuationToken can be used to retrieve them.
type AuditLogResult struct {
	DecoratedAuditLogEntries []AuditLogEntry `json:"decoratedAuditLogEntries"`
	ContinuationToken        string          `json:"continuationToken,omitempty"`
	HasMore                  bool            `json:"hasMore"`
}

// --- Audit operations ---

// QueryLog queries the organization audit log. startTime and endTime are
// optional ISO-8601 timestamps and batchSize optionally limits the number of
// entries per page. It returns the decorated audit log entries for the page;
// note that more pages may exist (see AuditLogResult.HasMore).
func (s *service) QueryLog(ctx context.Context, startTime, endTime string, batchSize int) ([]AuditLogEntry, error) {
	q := url.Values{}
	if startTime != "" {
		q.Set("startTime", startTime)
	}
	if endTime != "" {
		q.Set("endTime", endTime)
	}
	if batchSize > 0 {
		q.Set("batchSize", strconv.Itoa(batchSize))
	}
	var out AuditLogResult
	if err := s.c.Audit.GetJSON(ctx, "/_apis/audit/auditlog", q, &out); err != nil {
		return nil, err
	}
	return out.DecoratedAuditLogEntries, nil
}
