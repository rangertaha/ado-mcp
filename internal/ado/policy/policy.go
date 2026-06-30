// SPDX-License-Identifier: MIT

// Package policy exposes the Azure DevOps Policy service: project-scoped
// policy configurations (such as branch policies) and the policy types
// available to configure, including creating new configurations.
package policy

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "policy"

// service wraps the Azure DevOps clients for Policy operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// PolicyConfiguration is a configured policy instance within a project.
type PolicyConfiguration struct {
	ID         int    `json:"id"`
	Type       any    `json:"type,omitempty"`
	IsEnabled  bool   `json:"isEnabled"`
	IsBlocking bool   `json:"isBlocking"`
	Settings   any    `json:"settings,omitempty"`
	URL        string `json:"url,omitempty"`
}

// PolicyType describes a kind of policy that can be configured.
type PolicyType struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
}

// --- Operations ---

// ListConfigurations returns the policy configurations for a project.
func (s *service) ListConfigurations(ctx context.Context, project string) ([]PolicyConfiguration, error) {
	var out client.List[PolicyConfiguration]
	path := fmt.Sprintf("/%s/_apis/policy/configurations", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetConfiguration returns a single policy configuration by its ID.
func (s *service) GetConfiguration(ctx context.Context, project string, configurationID int) (*PolicyConfiguration, error) {
	var out PolicyConfiguration
	path := fmt.Sprintf("/%s/_apis/policy/configurations/%d", url.PathEscape(project), configurationID)
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTypes returns the policy types available within a project.
func (s *service) ListTypes(ctx context.Context, project string) ([]PolicyType, error) {
	var out client.List[PolicyType]
	path := fmt.Sprintf("/%s/_apis/policy/types", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// CreateConfiguration creates a new policy configuration within a project.
func (s *service) CreateConfiguration(ctx context.Context, project string, body map[string]any) (*PolicyConfiguration, error) {
	var out PolicyConfiguration
	path := fmt.Sprintf("/%s/_apis/policy/configurations", url.PathEscape(project))
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
