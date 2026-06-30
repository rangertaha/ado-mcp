// SPDX-License-Identifier: MIT

// Package advsec exposes the Azure DevOps Advanced Security service:
// querying code scanning alerts (code, secret and dependency) for a
// repository within a project.
package advsec

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "advsec"

// apiVersion is the preview API version required by the Advanced Security
// alert endpoints.
const apiVersion = "7.1-preview.1"

// service wraps the Azure DevOps clients for Advanced Security operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// Alert is a single Advanced Security alert raised against a repository.
type Alert struct {
	AlertID       int    `json:"alertId"`
	Title         string `json:"title,omitempty"`
	AlertType     string `json:"alertType,omitempty"`
	State         string `json:"state,omitempty"`
	Severity      string `json:"severity,omitempty"`
	RepositoryURL string `json:"repositoryUrl,omitempty"`
	FirstSeenDate string `json:"firstSeenDate,omitempty"`
	FixedDate     string `json:"fixedDate,omitempty"`
}

// --- Advanced Security operations ---

// ListAlerts lists Advanced Security alerts for a repository. alertType
// optionally filters by alert type (code, secret or dependency) and top
// optionally limits the number of alerts returned.
func (s *service) ListAlerts(ctx context.Context, project, repository, alertType string, top int) ([]Alert, error) {
	q := url.Values{}
	if alertType != "" {
		q.Set("criteria.alertType", alertType)
	}
	if top > 0 {
		q.Set("top", strconv.Itoa(top))
	}
	path := fmt.Sprintf("/%s/_apis/alert/repositories/%s/alerts",
		url.PathEscape(project), url.PathEscape(repository))
	var out client.List[Alert]
	if err := s.c.AdvSec.GetJSONVersion(ctx, path, q, &out, apiVersion); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetAlert returns a single Advanced Security alert by its numeric ID.
func (s *service) GetAlert(ctx context.Context, project, repository string, alertID int) (*Alert, error) {
	path := fmt.Sprintf("/%s/_apis/alert/repositories/%s/alerts/%d",
		url.PathEscape(project), url.PathEscape(repository), alertID)
	var out Alert
	if err := s.c.AdvSec.GetJSONVersion(ctx, path, nil, &out, apiVersion); err != nil {
		return nil, err
	}
	return &out, nil
}
