// SPDX-License-Identifier: MIT

// Package serviceendpoint provides read-only access to Azure DevOps service
// endpoints (service connections) and the available endpoint types.
package serviceendpoint

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name for the service endpoint area.
const Name = "serviceendpoint"

// service holds the Azure DevOps host clients used by this area.
type service struct{ c *ado.Clients }

// ServiceEndpoint represents an Azure DevOps service endpoint (service
// connection) used by pipelines and other features to authenticate to
// external services.
type ServiceEndpoint struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Type        string `json:"type,omitempty"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
	IsShared    bool   `json:"isShared,omitempty"`
	IsReady     bool   `json:"isReady,omitempty"`
}

// ServiceEndpointType describes a kind of service endpoint that can be created.
type ServiceEndpointType struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Description string `json:"description,omitempty"`
}

// ListEndpoints returns the service endpoints defined in the given project.
func (s *service) ListEndpoints(ctx context.Context, project string) ([]ServiceEndpoint, error) {
	var out client.List[ServiceEndpoint]
	path := fmt.Sprintf("/%s/_apis/serviceendpoint/endpoints", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetEndpoint returns a single service endpoint by its ID within a project.
func (s *service) GetEndpoint(ctx context.Context, project, endpointID string) (*ServiceEndpoint, error) {
	var out ServiceEndpoint
	path := fmt.Sprintf("/%s/_apis/serviceendpoint/endpoints/%s", url.PathEscape(project), url.PathEscape(endpointID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListTypes returns the available service endpoint types.
func (s *service) ListTypes(ctx context.Context) ([]ServiceEndpointType, error) {
	var out client.List[ServiceEndpointType]
	if err := s.c.Org.GetJSON(ctx, "/_apis/serviceendpoint/types", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
