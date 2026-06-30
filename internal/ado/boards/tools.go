// SPDX-License-Identifier: MIT

package boards

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the Boards tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "boards_list_backlogs", Title: "List backlogs",
		Description: "List the backlog levels configured for a team."}, svc.listBacklogs)
	server.Register(s, server.ToolDef{Name: "boards_list_backlog_work_items", Title: "List backlog work items",
		Description: "List the work items contained in a backlog level."}, svc.listBacklogWorkItems)
	server.Register(s, server.ToolDef{Name: "boards_list_iterations", Title: "List iterations",
		Description: "List the iterations (sprints) assigned to a team."}, svc.listIterations)
	server.Register(s, server.ToolDef{Name: "boards_get_iteration_capacity", Title: "Get iteration capacity",
		Description: "Get the team capacity for a specific iteration."}, svc.getIterationCapacity)
	server.Register(s, server.ToolDef{Name: "boards_list_boards", Title: "List boards",
		Description: "List the boards available to a team."}, svc.listBoards)
	server.Register(s, server.ToolDef{Name: "boards_get_board_columns", Title: "Get board columns",
		Description: "Get the columns configured on a board."}, svc.getBoardColumns)
	server.Register(s, server.ToolDef{Name: "boards_get_board_rows", Title: "Get board rows",
		Description: "Get the rows (swimlanes) configured on a board."}, svc.getBoardRows)
	server.Register(s, server.ToolDef{Name: "boards_list_board_charts", Title: "List board charts",
		Description: "List the charts available for a board."}, svc.listBoardCharts)
}

// ListBacklogsInput selects a team to list backlogs for.
type ListBacklogsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
}

func (s *service) listBacklogs(ctx context.Context, _ *mcp.CallToolRequest, in ListBacklogsInput) (*mcp.CallToolResult, server.ListResult[Backlog], error) {
	out, err := s.ListBacklogs(ctx, in.Project, in.Team)
	return nil, server.List(out), err
}

// ListBacklogWorkItemsInput selects a backlog level to list work items for.
type ListBacklogWorkItemsInput struct {
	Project   string `json:"project" jsonschema:"project name or ID"`
	Team      string `json:"team" jsonschema:"team name or ID"`
	BacklogID string `json:"backlogId" jsonschema:"backlog level ID"`
}

func (s *service) listBacklogWorkItems(ctx context.Context, _ *mcp.CallToolRequest, in ListBacklogWorkItemsInput) (*mcp.CallToolResult, server.ListResult[any], error) {
	out, err := s.ListBacklogWorkItems(ctx, in.Project, in.Team, in.BacklogID)
	return nil, server.List(out), err
}

// ListIterationsInput selects a team to list iterations for.
type ListIterationsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
}

func (s *service) listIterations(ctx context.Context, _ *mcp.CallToolRequest, in ListIterationsInput) (*mcp.CallToolResult, server.ListResult[Iteration], error) {
	out, err := s.ListIterations(ctx, in.Project, in.Team)
	return nil, server.List(out), err
}

// GetIterationCapacityInput selects an iteration to read capacity for.
type GetIterationCapacityInput struct {
	Project     string `json:"project" jsonschema:"project name or ID"`
	Team        string `json:"team" jsonschema:"team name or ID"`
	IterationID string `json:"iterationId" jsonschema:"iteration ID"`
}

func (s *service) getIterationCapacity(ctx context.Context, _ *mcp.CallToolRequest, in GetIterationCapacityInput) (*mcp.CallToolResult, server.ListResult[Capacity], error) {
	out, err := s.GetIterationCapacity(ctx, in.Project, in.Team, in.IterationID)
	return nil, server.List(out), err
}

// ListBoardsInput selects a team to list boards for.
type ListBoardsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
}

func (s *service) listBoards(ctx context.Context, _ *mcp.CallToolRequest, in ListBoardsInput) (*mcp.CallToolResult, server.ListResult[Board], error) {
	out, err := s.ListBoards(ctx, in.Project, in.Team)
	return nil, server.List(out), err
}

// GetBoardColumnsInput selects a board to read columns for.
type GetBoardColumnsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
	Board   string `json:"board" jsonschema:"board name or ID"`
}

func (s *service) getBoardColumns(ctx context.Context, _ *mcp.CallToolRequest, in GetBoardColumnsInput) (*mcp.CallToolResult, server.ListResult[BoardColumn], error) {
	out, err := s.GetBoardColumns(ctx, in.Project, in.Team, in.Board)
	return nil, server.List(out), err
}

// GetBoardRowsInput selects a board to read rows for.
type GetBoardRowsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
	Board   string `json:"board" jsonschema:"board name or ID"`
}

func (s *service) getBoardRows(ctx context.Context, _ *mcp.CallToolRequest, in GetBoardRowsInput) (*mcp.CallToolResult, server.ListResult[BoardRow], error) {
	out, err := s.GetBoardRows(ctx, in.Project, in.Team, in.Board)
	return nil, server.List(out), err
}

// ListBoardChartsInput selects a board to list charts for.
type ListBoardChartsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Team    string `json:"team" jsonschema:"team name or ID"`
	Board   string `json:"board" jsonschema:"board name or ID"`
}

func (s *service) listBoardCharts(ctx context.Context, _ *mcp.CallToolRequest, in ListBoardChartsInput) (*mcp.CallToolResult, server.ListResult[BoardChart], error) {
	out, err := s.ListBoardCharts(ctx, in.Project, in.Team, in.Board)
	return nil, server.List(out), err
}
