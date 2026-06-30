// SPDX-License-Identifier: MIT

// Package graph exposes the Azure DevOps Graph/Identity service:
// users, groups and subject memberships. All operations are read-only and
// reach the vssps.dev.azure.com host using preview API endpoints.
package graph

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "graph"

// apiVersion is the preview API version required by the Graph endpoints.
const apiVersion = "7.1-preview.1"

// service wraps the Azure DevOps clients for Graph/Identity operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// GraphUser is a user subject in the organization's graph.
type GraphUser struct {
	SubjectKind   string `json:"subjectKind,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	PrincipalName string `json:"principalName,omitempty"`
	MailAddress   string `json:"mailAddress,omitempty"`
	Descriptor    string `json:"descriptor,omitempty"`
	Origin        string `json:"origin,omitempty"`
}

// GraphGroup is a group subject in the organization's graph.
type GraphGroup struct {
	SubjectKind   string `json:"subjectKind,omitempty"`
	DisplayName   string `json:"displayName,omitempty"`
	PrincipalName string `json:"principalName,omitempty"`
	Description   string `json:"description,omitempty"`
	Descriptor    string `json:"descriptor,omitempty"`
}

// Membership is a relationship between a member subject and its container.
type Membership struct {
	ContainerDescriptor string `json:"containerDescriptor,omitempty"`
	MemberDescriptor    string `json:"memberDescriptor,omitempty"`
}

// Descriptor resolves a storage key to its subject descriptor value.
type Descriptor struct {
	Value string `json:"value,omitempty"`
}

// --- Operations ---

// ListUsers lists the user subjects in the organization graph.
func (s *service) ListUsers(ctx context.Context) ([]GraphUser, error) {
	var out client.List[GraphUser]
	if err := s.c.VSSPS.GetJSONVersion(ctx, "/_apis/graph/users", nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListGroups lists the group subjects in the organization graph.
func (s *service) ListGroups(ctx context.Context) ([]GraphGroup, error) {
	var out client.List[GraphGroup]
	if err := s.c.VSSPS.GetJSONVersion(ctx, "/_apis/graph/groups", nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListMemberships lists the memberships of a subject. The optional direction
// ("up" or "down") controls whether containers or members are returned.
func (s *service) ListMemberships(ctx context.Context, subjectDescriptor, direction string) ([]Membership, error) {
	q := url.Values{}
	if direction != "" {
		q.Set("direction", direction)
	}
	var out client.List[Membership]
	path := fmt.Sprintf("/_apis/graph/Memberships/%s", url.PathEscape(subjectDescriptor))
	if err := s.c.VSSPS.GetJSONVersion(ctx, path, q, &out, apiVersion); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// AddMembership creates a membership relating the subject to the container.
func (s *service) AddMembership(ctx context.Context, subjectDescriptor, containerDescriptor string) (*Membership, error) {
	var m Membership
	path := fmt.Sprintf("/_apis/graph/memberships/%s/%s",
		url.PathEscape(subjectDescriptor), url.PathEscape(containerDescriptor))
	if _, err := s.c.VSSPS.Do(ctx, client.Request{
		Method:     "PUT",
		Path:       path,
		APIVersion: apiVersion,
		Out:        &m,
	}); err != nil {
		return nil, err
	}
	return &m, nil
}

// RemoveMembership removes the membership relating the subject to the container.
func (s *service) RemoveMembership(ctx context.Context, subjectDescriptor, containerDescriptor string) error {
	path := fmt.Sprintf("/_apis/graph/memberships/%s/%s",
		url.PathEscape(subjectDescriptor), url.PathEscape(containerDescriptor))
	_, err := s.c.VSSPS.Do(ctx, client.Request{
		Method:     "DELETE",
		Path:       path,
		APIVersion: apiVersion,
	})
	return err
}

// GetDescriptor resolves a storage key (subject id) to its graph descriptor.
func (s *service) GetDescriptor(ctx context.Context, storageKey string) (*Descriptor, error) {
	var d Descriptor
	path := fmt.Sprintf("/_apis/graph/descriptors/%s", url.PathEscape(storageKey))
	if err := s.c.VSSPS.GetJSONVersion(ctx, path, nil, &d, apiVersion); err != nil {
		return nil, err
	}
	return &d, nil
}
