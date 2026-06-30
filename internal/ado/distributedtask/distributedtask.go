// SPDX-License-Identifier: MIT

// Package distributedtask exposes the Azure DevOps Distributed Task service:
// pipeline resources such as variable groups, environments, agent pools and
// secure files, including listing, fetching and creating selected resources.
package distributedtask

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "distributedtask"

// service wraps the Azure DevOps clients for Distributed Task operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// VariableGroup is a named collection of pipeline variables.
type VariableGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
	Variables   any    `json:"variables,omitempty"`
}

// Environment is a pipeline deployment environment.
type Environment struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AgentPool is an organization-level pool of build/release agents.
type AgentPool struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	IsHosted bool   `json:"isHosted,omitempty"`
	PoolType string `json:"poolType,omitempty"`
}

// SecureFile is a secure file stored for use by pipelines.
type SecureFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// --- Operations ---

// ListVariableGroups returns the variable groups within a project.
func (s *service) ListVariableGroups(ctx context.Context, project string) ([]VariableGroup, error) {
	var out client.List[VariableGroup]
	path := fmt.Sprintf("/%s/_apis/distributedtask/variablegroups", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetVariableGroup returns a single variable group by its ID within a project.
func (s *service) GetVariableGroup(ctx context.Context, project string, groupID int) (*VariableGroup, error) {
	var out VariableGroup
	path := fmt.Sprintf("/%s/_apis/distributedtask/variablegroups/%s",
		url.PathEscape(project), url.PathEscape(strconv.Itoa(groupID)))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateVariableGroup creates a variable group within a project from a raw body.
func (s *service) CreateVariableGroup(ctx context.Context, project string, body map[string]any) (*VariableGroup, error) {
	var out VariableGroup
	path := fmt.Sprintf("/%s/_apis/distributedtask/variablegroups", url.PathEscape(project))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEnvironments returns the deployment environments within a project.
func (s *service) ListEnvironments(ctx context.Context, project string) ([]Environment, error) {
	var out client.List[Environment]
	path := fmt.Sprintf("/%s/_apis/distributedtask/environments", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetEnvironment returns a single deployment environment by its ID within a project.
func (s *service) GetEnvironment(ctx context.Context, project string, environmentID int) (*Environment, error) {
	var out Environment
	path := fmt.Sprintf("/%s/_apis/distributedtask/environments/%s",
		url.PathEscape(project), url.PathEscape(strconv.Itoa(environmentID)))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateEnvironment creates a deployment environment within a project.
func (s *service) CreateEnvironment(ctx context.Context, project, name, description string) (*Environment, error) {
	var out Environment
	path := fmt.Sprintf("/%s/_apis/distributedtask/environments", url.PathEscape(project))
	body := map[string]string{"name": name, "description": description}
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListAgentPools returns the organization-level agent pools.
func (s *service) ListAgentPools(ctx context.Context) ([]AgentPool, error) {
	var out client.List[AgentPool]
	if err := s.c.Org.GetJSON(ctx, "/_apis/distributedtask/pools", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListSecureFiles returns the secure files within a project.
func (s *service) ListSecureFiles(ctx context.Context, project string) ([]SecureFile, error) {
	var out client.List[SecureFile]
	path := fmt.Sprintf("/%s/_apis/distributedtask/securefiles", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
