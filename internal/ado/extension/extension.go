// SPDX-License-Identifier: MIT

// Package extension exposes the Azure DevOps Extension Management service:
// listing and retrieving extensions installed in the organization.
package extension

import (
	"context"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "extension"

// apiVersion is the preview API version required by the extension
// management endpoints.
const apiVersion = "7.1-preview.1"

// service wraps the Azure DevOps clients for Extension Management operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// InstalledExtension describes an extension installed in the organization.
type InstalledExtension struct {
	ExtensionID   string `json:"extensionId,omitempty"`
	ExtensionName string `json:"extensionName,omitempty"`
	PublisherID   string `json:"publisherId,omitempty"`
	PublisherName string `json:"publisherName,omitempty"`
	Version       string `json:"version,omitempty"`
	Flags         any    `json:"flags,omitempty"`
	InstallState  any    `json:"installState,omitempty"`
}

// --- Extension operations ---

// ListInstalled returns all extensions installed in the organization. The
// endpoint responds with the standard {"count":N,"value":[...]} envelope.
func (s *service) ListInstalled(ctx context.Context) ([]InstalledExtension, error) {
	var out client.List[InstalledExtension]
	if err := s.c.ExtMgmt.GetJSONVersion(ctx, "/_apis/extensionmanagement/installedextensions", nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// Get retrieves a single installed extension by its publisher and extension
// identifiers.
func (s *service) Get(ctx context.Context, publisherID, extensionID string) (*InstalledExtension, error) {
	path := "/_apis/extensionmanagement/installedextensionsbyname/" +
		url.PathEscape(publisherID) + "/" + url.PathEscape(extensionID)
	var out InstalledExtension
	if err := s.c.ExtMgmt.GetJSONVersion(ctx, path, nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return &out, nil
}
