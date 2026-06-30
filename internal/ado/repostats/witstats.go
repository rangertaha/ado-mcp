// SPDX-License-Identifier: MIT

package repostats

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rangertaha/ado-mcp/internal/client"
)

// defaultMaxItems bounds how many work items a stats call will fetch.
const defaultMaxItems = 2000

// witBatchSize is the maximum IDs per work-item batch request.
const witBatchSize = 200

type wiqlResult struct {
	WorkItems []struct {
		ID int `json:"id"`
	} `json:"workItems"`
}

type witFields struct {
	Fields map[string]any `json:"fields"`
}

// WorkItemStats summarizes work items broken down by type, state, and assignee.
type WorkItemStats struct {
	Scope        string         `json:"scope" jsonschema:"what was analyzed"`
	ItemsScanned int            `json:"itemsScanned" jsonschema:"work items examined"`
	Truncated    bool           `json:"truncated" jsonschema:"true if the scan limit was hit"`
	ByType       map[string]int `json:"byType" jsonschema:"counts keyed by work item type"`
	ByState      map[string]int `json:"byState" jsonschema:"counts keyed by state"`
	ByAssignee   map[string]int `json:"byAssignee" jsonschema:"counts keyed by assignee display name"`
}

// fieldString extracts a string field, resolving identity objects to their
// displayName.
func fieldString(fields map[string]any, name string) string {
	v, ok := fields[name]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case map[string]any:
		if dn, ok := t["displayName"].(string); ok {
			return dn
		}
	}
	return fmt.Sprintf("%v", v)
}

// WorkItemStatsForQuery runs a WIQL query (defaulting to all work items in the
// project) and tallies the matching items by type, state, and assignee.
func (s *service) WorkItemStatsForQuery(ctx context.Context, project, wiql string, maxItems int) (*WorkItemStats, error) {
	if maxItems <= 0 {
		maxItems = defaultMaxItems
	}
	if strings.TrimSpace(wiql) == "" {
		wiql = "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project ORDER BY [System.ChangedDate] DESC"
	}

	var qr wiqlResult
	wiqlPath := fmt.Sprintf("/%s/_apis/wit/wiql", url.PathEscape(project))
	q := url.Values{}
	// Request one extra so we can tell a full result set apart from a capped one.
	q.Set("$top", strconv.Itoa(maxItems+1))
	if err := s.c.Org.PostJSON(ctx, wiqlPath, q, map[string]string{"query": wiql}, &qr); err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(qr.WorkItems))
	for _, w := range qr.WorkItems {
		ids = append(ids, w.ID)
	}
	truncated := false
	if len(ids) > maxItems {
		ids = ids[:maxItems]
		truncated = true
	}

	st := &WorkItemStats{
		Scope:      "project:" + project,
		ByType:     map[string]int{},
		ByState:    map[string]int{},
		ByAssignee: map[string]int{},
	}

	for start := 0; start < len(ids); start += witBatchSize {
		end := start + witBatchSize
		if end > len(ids) {
			end = len(ids)
		}
		batch := ids[start:end]
		strIDs := make([]string, len(batch))
		for i, id := range batch {
			strIDs[i] = strconv.Itoa(id)
		}
		bq := url.Values{}
		bq.Set("ids", strings.Join(strIDs, ","))
		bq.Set("fields", "System.WorkItemType,System.State,System.AssignedTo")
		var page client.List[witFields]
		if err := s.c.Org.GetJSON(ctx, "/_apis/wit/workitems", bq, &page); err != nil {
			return nil, err
		}
		for _, wi := range page.Value {
			st.ItemsScanned++
			if t := fieldString(wi.Fields, "System.WorkItemType"); t != "" {
				st.ByType[t]++
			}
			if state := fieldString(wi.Fields, "System.State"); state != "" {
				st.ByState[state]++
			}
			assignee := fieldString(wi.Fields, "System.AssignedTo")
			if assignee == "" {
				assignee = "(unassigned)"
			}
			st.ByAssignee[assignee]++
		}
	}
	st.Truncated = truncated
	return st, nil
}
