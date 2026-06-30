// SPDX-License-Identifier: MIT

// Package security provides read-only access to Azure DevOps security
// namespaces and access control lists at the organization level.
package security

import (
	"context"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name for this area.
const Name = "security"

// service holds the Azure DevOps client used by the security area.
type service struct {
	c *ado.Clients
}

// SecurityNamespace describes a security namespace, which groups a set of
// related permissions (actions) under a stable namespace identifier.
type SecurityNamespace struct {
	NamespaceID string `json:"namespaceId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	Actions     any    `json:"actions,omitempty"`
}

// AccessControlList is the set of access control entries (ACEs) that apply to
// a single token within a security namespace.
type AccessControlList struct {
	InheritPermissions bool   `json:"inheritPermissions"`
	Token              string `json:"token,omitempty"`
	AcesDictionary     any    `json:"acesDictionary,omitempty"`
}

// ListNamespaces returns the security namespaces in the organization,
// optionally filtered to a single namespace by ID.
func (s *service) ListNamespaces(ctx context.Context, namespaceID string) ([]SecurityNamespace, error) {
	q := url.Values{}
	if namespaceID != "" {
		q.Set("namespaceId", namespaceID)
	}
	var out client.List[SecurityNamespace]
	if err := s.c.Org.GetJSON(ctx, "/_apis/securitynamespaces", q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListAccessControlLists returns the access control lists for a security
// namespace, optionally filtered by token and including extended info.
func (s *service) ListAccessControlLists(ctx context.Context, namespaceID, token string, includeExtendedInfo bool) ([]AccessControlList, error) {
	q := url.Values{}
	if token != "" {
		q.Set("token", token)
	}
	if includeExtendedInfo {
		q.Set("includeExtendedInfo", strconv.FormatBool(includeExtendedInfo))
	}
	var out client.List[AccessControlList]
	path := "/_apis/accesscontrollists/" + url.PathEscape(namespaceID)
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// SetAccessControlEntries adds or updates access control entries on the
// access control list for a token in the given security namespace. The body
// is passed through to the API and typically contains the token, a merge
// flag, and the access control entries to apply.
func (s *service) SetAccessControlEntries(ctx context.Context, namespaceID string, body map[string]any) ([]any, error) {
	var out client.List[any]
	path := "/_apis/accesscontrolentries/" + url.PathEscape(namespaceID)
	if err := s.c.Org.PostJSON(ctx, path, nil, body, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// RemoveAccessControlLists removes the access control lists for the given
// tokens within a security namespace.
func (s *service) RemoveAccessControlLists(ctx context.Context, namespaceID, tokens string) error {
	q := url.Values{}
	if tokens != "" {
		q.Set("tokens", tokens)
	}
	path := "/_apis/accesscontrollists/" + url.PathEscape(namespaceID)
	return s.c.Org.Delete(ctx, path, q, nil)
}
