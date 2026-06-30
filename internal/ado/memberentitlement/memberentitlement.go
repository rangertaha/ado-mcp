// SPDX-License-Identifier: MIT

// Package memberentitlement exposes the Azure DevOps Member Entitlement
// Management service: listing and inspecting user and group access-level
// entitlements (licensing) for an organization.
package memberentitlement

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "memberentitlement"

// API versions required by the member entitlement endpoints. They differ per
// route: user entitlements accept 7.1-preview.3 while group entitlements only
// accept 7.1-preview.1.
const (
	apiVersion      = "7.1-preview.3"
	groupAPIVersion = "7.1-preview.1"
)

// service wraps the Azure DevOps clients for member entitlement operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// UserEntitlement describes a user's access level and licensing within the
// organization.
type UserEntitlement struct {
	ID               string `json:"id"`
	User             any    `json:"user,omitempty"`
	AccessLevel      any    `json:"accessLevel,omitempty"`
	LastAccessedDate string `json:"lastAccessedDate,omitempty"`
	DateCreated      string `json:"dateCreated,omitempty"`
}

// UserEntitlementList is the response envelope returned when listing user
// entitlements. It is a JSON object wrapping the page of members.
type UserEntitlementList struct {
	Members           []UserEntitlement `json:"members"`
	ContinuationToken any               `json:"continuationToken,omitempty"`
	TotalCount        int               `json:"totalCount"`
}

// GroupEntitlement describes a group's licensing rule within the organization.
type GroupEntitlement struct {
	ID          string `json:"id"`
	Group       any    `json:"group,omitempty"`
	LicenseRule any    `json:"licenseRule,omitempty"`
}

// --- Member entitlement operations ---

// ListUsers returns a page of user entitlements for the organization.
func (s *service) ListUsers(ctx context.Context) (*UserEntitlementList, error) {
	var out UserEntitlementList
	if err := s.c.VSAEX.GetJSONVersion(ctx, "/_apis/userentitlements", nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUser returns the entitlement for a single user by their entitlement id.
func (s *service) GetUser(ctx context.Context, userID string) (*UserEntitlement, error) {
	path := "/_apis/userentitlements/" + url.PathEscape(userID)
	var out UserEntitlement
	if err := s.c.VSAEX.GetJSONVersion(ctx, path, nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListGroups returns the group entitlements for the organization. Depending on
// the API version the endpoint returns either a bare array or an envelope
// ({"members":[...]} or {"value":[...]}); this handles all three.
func (s *service) ListGroups(ctx context.Context) ([]GroupEntitlement, error) {
	var raw json.RawMessage
	if err := s.c.VSAEX.GetJSONVersion(ctx, "/_apis/groupentitlements", nil, &raw, groupAPIVersion); err != nil {
		return nil, err
	}
	// Try a bare array first, then the common envelopes.
	var arr []GroupEntitlement
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	var env struct {
		Members []GroupEntitlement `json:"members"`
		Value   []GroupEntitlement `json:"value"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decoding group entitlements: %w", err)
	}
	if env.Members != nil {
		return env.Members, nil
	}
	return env.Value, nil
}
