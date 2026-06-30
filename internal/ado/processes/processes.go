// SPDX-License-Identifier: MIT

// Package processes exposes the Azure DevOps Work Item Tracking process
// customization service: organization processes and their work item types,
// fields, states, and behaviors.
package processes

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "processes"

// service wraps the Azure DevOps clients for process operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Process is an organization work item tracking process.
type Process struct {
	TypeID              string `json:"typeId"`
	Name                string `json:"name"`
	ReferenceName       string `json:"referenceName,omitempty"`
	Description         string `json:"description,omitempty"`
	IsEnabled           bool   `json:"isEnabled,omitempty"`
	IsDefault           bool   `json:"isDefault,omitempty"`
	ParentProcessTypeID string `json:"parentProcessTypeId,omitempty"`
}

// ProcessWorkItemType is a work item type defined within a process.
type ProcessWorkItemType struct {
	ReferenceName string `json:"referenceName"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Color         string `json:"color,omitempty"`
	Icon          string `json:"icon,omitempty"`
	IsDisabled    bool   `json:"isDisabled,omitempty"`
	Customization string `json:"customization,omitempty"`
}

// ProcessField is a field on a process work item type.
type ProcessField struct {
	ReferenceName string `json:"referenceName"`
	Name          string `json:"name"`
	Type          string `json:"type,omitempty"`
	Required      bool   `json:"required,omitempty"`
	ReadOnly      bool   `json:"readOnly,omitempty"`
}

// ProcessState is a workflow state of a process work item type.
type ProcessState struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Color         string `json:"color,omitempty"`
	StateCategory string `json:"stateCategory,omitempty"`
	Order         int    `json:"order,omitempty"`
}

// Behavior is a behavior defined within a process.
type Behavior struct {
	ReferenceName string `json:"referenceName"`
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Rank          int    `json:"rank,omitempty"`
	Color         string `json:"color,omitempty"`
}

// --- Operations ---

// List returns the work item tracking processes in the organization.
func (s *service) List(ctx context.Context) ([]Process, error) {
	var out client.List[Process]
	if err := s.c.Org.GetJSON(ctx, "/_apis/work/processes", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// Get returns a single process by its process type ID.
func (s *service) Get(ctx context.Context, processTypeID string) (*Process, error) {
	var out Process
	path := fmt.Sprintf("/_apis/work/processes/%s", url.PathEscape(processTypeID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListWorkItemTypes returns the work item types defined within a process.
func (s *service) ListWorkItemTypes(ctx context.Context, processTypeID string) ([]ProcessWorkItemType, error) {
	var out client.List[ProcessWorkItemType]
	path := fmt.Sprintf("/_apis/work/processes/%s/workitemtypes", url.PathEscape(processTypeID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListFields returns the fields of a work item type within a process.
func (s *service) ListFields(ctx context.Context, processTypeID, witRefName string) ([]ProcessField, error) {
	var out client.List[ProcessField]
	path := fmt.Sprintf("/_apis/work/processes/%s/workItemTypes/%s/fields",
		url.PathEscape(processTypeID), url.PathEscape(witRefName))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListStates returns the workflow states of a work item type within a process.
func (s *service) ListStates(ctx context.Context, processTypeID, witRefName string) ([]ProcessState, error) {
	var out client.List[ProcessState]
	path := fmt.Sprintf("/_apis/work/processes/%s/workItemTypes/%s/states",
		url.PathEscape(processTypeID), url.PathEscape(witRefName))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListBehaviors returns the behaviors defined within a process.
func (s *service) ListBehaviors(ctx context.Context, processTypeID string) ([]Behavior, error) {
	var out client.List[Behavior]
	path := fmt.Sprintf("/_apis/work/processes/%s/behaviors", url.PathEscape(processTypeID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
