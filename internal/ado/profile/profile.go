// SPDX-License-Identifier: MIT

// Package profile exposes the Azure DevOps Profile service: retrieving the
// profile of the authenticated user.
package profile

import (
	"context"

	"github.com/rangertaha/ado-mcp/internal/ado"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "profile"

// service wraps the Azure DevOps clients for Profile operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Profile is the profile of an Azure DevOps user.
type Profile struct {
	ID           string `json:"id"`
	DisplayName  string `json:"displayName,omitempty"`
	EmailAddress string `json:"emailAddress,omitempty"`
	PublicAlias  string `json:"publicAlias,omitempty"`
	CoreRevision int    `json:"coreRevision,omitempty"`
	TimeStamp    string `json:"timeStamp,omitempty"`
	Revision     int    `json:"revision,omitempty"`
}

// --- Profile operations ---

// Check verifies connectivity and credentials by fetching the authenticated
// user's profile. It is used by the `ado test` command.
func Check(ctx context.Context, c *ado.Clients) (*Profile, error) {
	return (&service{c: c}).GetMe(ctx)
}

// GetMe retrieves the profile of the authenticated user.
func (s *service) GetMe(ctx context.Context) (*Profile, error) {
	var out Profile
	if err := s.c.VSSPS.GetJSONVersion(ctx, "/_apis/profile/profiles/me", nil, &out, "7.1-preview.3"); err != nil {
		return nil, err
	}
	return &out, nil
}
