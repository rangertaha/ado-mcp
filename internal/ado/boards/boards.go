// SPDX-License-Identifier: MIT

// Package boards exposes Azure DevOps Boards / Work tracking endpoints
// (backlogs, iterations, team capacity, and boards/columns). All operations
// are team-scoped and read-only, and use the primary organization host.
package boards

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset identifier for this area.
const Name = "boards"

// service wires the shared Azure DevOps clients into the Boards operations.
type service struct{ c *ado.Clients }

// Backlog describes a backlog level configured for a team.
type Backlog struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Rank               int    `json:"rank,omitempty"`
	WorkItemCountLimit int    `json:"workItemCountLimit,omitempty"`
}

// Iteration describes a team iteration (sprint).
type Iteration struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path,omitempty"`
	Attributes any    `json:"attributes,omitempty"`
}

// Capacity describes a team member's capacity within an iteration.
type Capacity struct {
	TeamMember any `json:"teamMember,omitempty"`
	Activities any `json:"activities,omitempty"`
}

// Board describes a team board.
type Board struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// BoardColumn describes a single column on a board.
type BoardColumn struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ItemLimit  int    `json:"itemLimit,omitempty"`
	ColumnType string `json:"columnType,omitempty"`
}

// BoardRow describes a single row (swimlane) on a board.
type BoardRow struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BoardChart describes a chart available for a board.
type BoardChart struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// ListBacklogs lists the backlog levels configured for a team.
func (s *service) ListBacklogs(ctx context.Context, project, team string) ([]Backlog, error) {
	var out client.List[Backlog]
	path := fmt.Sprintf("/%s/%s/_apis/work/backlogs", url.PathEscape(project), url.PathEscape(team))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListBacklogWorkItems lists the work items contained in a backlog level.
func (s *service) ListBacklogWorkItems(ctx context.Context, project, team, backlogID string) ([]any, error) {
	var out client.List[any]
	path := fmt.Sprintf("/%s/%s/_apis/work/backlogs/%s/workItems",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(backlogID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListIterations lists the iterations assigned to a team.
func (s *service) ListIterations(ctx context.Context, project, team string) ([]Iteration, error) {
	var out client.List[Iteration]
	path := fmt.Sprintf("/%s/%s/_apis/work/teamsettings/iterations", url.PathEscape(project), url.PathEscape(team))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetIterationCapacity returns the team capacity for a specific iteration.
func (s *service) GetIterationCapacity(ctx context.Context, project, team, iterationID string) ([]Capacity, error) {
	var out client.List[Capacity]
	path := fmt.Sprintf("/%s/%s/_apis/work/teamsettings/iterations/%s/capacities",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(iterationID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListBoards lists the boards available to a team.
func (s *service) ListBoards(ctx context.Context, project, team string) ([]Board, error) {
	var out client.List[Board]
	path := fmt.Sprintf("/%s/%s/_apis/work/boards", url.PathEscape(project), url.PathEscape(team))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetBoardColumns returns the columns configured on a board.
func (s *service) GetBoardColumns(ctx context.Context, project, team, board string) ([]BoardColumn, error) {
	var out client.List[BoardColumn]
	path := fmt.Sprintf("/%s/%s/_apis/work/boards/%s/columns",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(board))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetBoardRows returns the rows (swimlanes) configured on a board.
func (s *service) GetBoardRows(ctx context.Context, project, team, board string) ([]BoardRow, error) {
	var out client.List[BoardRow]
	path := fmt.Sprintf("/%s/%s/_apis/work/boards/%s/rows",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(board))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListBoardCharts lists the charts available for a board.
func (s *service) ListBoardCharts(ctx context.Context, project, team, board string) ([]BoardChart, error) {
	var out client.List[BoardChart]
	path := fmt.Sprintf("/%s/%s/_apis/work/boards/%s/charts",
		url.PathEscape(project), url.PathEscape(team), url.PathEscape(board))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
