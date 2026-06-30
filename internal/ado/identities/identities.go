// SPDX-License-Identifier: MIT

// Package identities exposes the Azure DevOps legacy Identity service for
// looking up identities (users and groups) by search filter or descriptor.
// All operations are read-only and reach the vssps.dev.azure.com host.
package identities

import (
	"context"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "identities"

// service wraps the Azure DevOps clients for legacy identity operations.
type service struct {
	c *ado.Clients
}

// Identity is a legacy identity record describing a user or group subject.
type Identity struct {
	ID                  string `json:"id,omitempty"`
	Descriptor          string `json:"descriptor,omitempty"`
	SubjectDescriptor   string `json:"subjectDescriptor,omitempty"`
	ProviderDisplayName string `json:"providerDisplayName,omitempty"`
	IsActive            bool   `json:"isActive,omitempty"`
	Members             any    `json:"members,omitempty"`
	MemberOf            any    `json:"memberOf,omitempty"`
	Properties          any    `json:"properties,omitempty"`
}

// Read looks up identities using an optional search filter, filter value and
// comma-separated list of descriptors.
func (s *service) Read(ctx context.Context, searchFilter, filterValue, descriptors string) ([]Identity, error) {
	q := url.Values{}
	if searchFilter != "" {
		q.Set("searchFilter", searchFilter)
	}
	if filterValue != "" {
		q.Set("filterValue", filterValue)
	}
	if descriptors != "" {
		q.Set("descriptors", descriptors)
	}
	var out client.List[Identity]
	if err := s.c.VSSPS.GetJSON(ctx, "/_apis/identities", q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
