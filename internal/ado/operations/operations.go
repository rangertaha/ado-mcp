// SPDX-License-Identifier: MIT

// Package operations exposes the Azure DevOps Operations service: querying the
// status of long-running asynchronous operations by their operation ID.
package operations

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "operations"

// service wraps the Azure DevOps clients for Operations operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Operation describes the status of a long-running asynchronous operation.
type Operation struct {
	ID              string `json:"id"`
	Status          string `json:"status,omitempty"`
	URL             string `json:"url,omitempty"`
	ResultMessage   string `json:"resultMessage,omitempty"`
	DetailedMessage string `json:"detailedMessage,omitempty"`
}

// --- Operations operations ---

// Get retrieves the status of a single asynchronous operation by its ID.
func (s *service) Get(ctx context.Context, operationID string) (*Operation, error) {
	path := fmt.Sprintf("/_apis/operations/%s", url.PathEscape(operationID))
	var out Operation
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
