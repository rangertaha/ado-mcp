// SPDX-License-Identifier: MIT

// Package ado holds the connection to an Azure DevOps organization that the
// per-area tool packages (core, wit, git, …) share.
//
// Azure DevOps spreads its APIs across several hosts. Clients bundles a
// configured REST client for each host an area might need:
//
//	dev.azure.com/{org}            primary services (core, wit, git, build, …)
//	vssps.dev.azure.com/{org}      Graph/Identity
//	vsrm.dev.azure.com/{org}       Release Management
//	feeds.dev.azure.com/{org}      Artifacts/Packaging feeds
//	auditservice.dev.azure.com/{org} Audit log
package ado

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/rangertaha/ado-mcp/internal/client"
)

// DefaultAPIVersion is the Azure DevOps REST API version targeted by default.
// Individual requests may override it (e.g. for "-preview" endpoints).
const DefaultAPIVersion = "7.1"

// Clients bundles the REST clients needed to reach the full Azure DevOps API.
type Clients struct {
	// Org reaches the primary organization host, https://dev.azure.com/{org}.
	Org *client.Client
	// VSSPS reaches the Graph/Identity host, https://vssps.dev.azure.com/{org}.
	VSSPS *client.Client
	// VSRM reaches the Release Management host, https://vsrm.dev.azure.com/{org}.
	VSRM *client.Client
	// Feeds reaches the Artifacts host, https://feeds.dev.azure.com/{org}.
	Feeds *client.Client
	// Audit reaches the Audit host, https://auditservice.dev.azure.com/{org}.
	Audit *client.Client
	// Search reaches the Search host, https://almsearch.dev.azure.com/{org}.
	Search *client.Client
	// VSAEX reaches the Member Entitlement host, https://vsaex.dev.azure.com/{org}.
	VSAEX *client.Client
	// ExtMgmt reaches the Extension Management host, https://extmgmt.dev.azure.com/{org}.
	ExtMgmt *client.Client
	// AdvSec reaches the Advanced Security host, https://advsec.dev.azure.com/{org}.
	AdvSec *client.Client
}

// NewClients builds the Azure DevOps clients for the given organization URL
// (e.g. "https://dev.azure.com/myorg") authenticated with a PAT.
func NewClients(orgURL, pat string, opts ...client.Option) (*Clients, error) {
	auth := client.NewPATAuthorizer(pat)
	base := append([]client.Option{
		client.WithAPIVersion(DefaultAPIVersion),
		client.WithUserAgent("ado-mcp"),
	}, opts...)

	mk := func(label, subdomain string) (*client.Client, error) {
		u, err := hostURL(orgURL, subdomain)
		if err != nil {
			return nil, err
		}
		c, err := client.New(u, auth, base...)
		if err != nil {
			return nil, fmt.Errorf("creating %s client: %w", label, err)
		}
		return c, nil
	}

	org, err := mk("org", "")
	if err != nil {
		return nil, err
	}
	vssps, err := mk("vssps", "vssps")
	if err != nil {
		return nil, err
	}
	vsrm, err := mk("vsrm", "vsrm")
	if err != nil {
		return nil, err
	}
	feeds, err := mk("feeds", "feeds")
	if err != nil {
		return nil, err
	}
	audit, err := mk("audit", "auditservice")
	if err != nil {
		return nil, err
	}
	search, err := mk("search", "almsearch")
	if err != nil {
		return nil, err
	}
	vsaex, err := mk("vsaex", "vsaex")
	if err != nil {
		return nil, err
	}
	extmgmt, err := mk("extmgmt", "extmgmt")
	if err != nil {
		return nil, err
	}
	advsec, err := mk("advsec", "advsec")
	if err != nil {
		return nil, err
	}

	return &Clients{
		Org: org, VSSPS: vssps, VSRM: vsrm, Feeds: feeds, Audit: audit, Search: search,
		VSAEX: vsaex, ExtMgmt: extmgmt, AdvSec: advsec,
	}, nil
}

// hostURL derives a service host URL from the organization URL by inserting the
// given subdomain. An empty subdomain returns the org URL unchanged.
//
//	https://dev.azure.com/myorg      + "vsrm"  -> https://vsrm.dev.azure.com/myorg
//	https://myorg.visualstudio.com   + "vsrm"  -> https://myorg.vsrm.visualstudio.com
func hostURL(orgURL, subdomain string) (string, error) {
	u, err := url.Parse(orgURL)
	if err != nil {
		return "", fmt.Errorf("invalid org URL %q: %w", orgURL, err)
	}
	if subdomain == "" {
		return u.String(), nil
	}
	switch {
	case u.Host == "dev.azure.com":
		u.Host = subdomain + ".dev.azure.com"
	case strings.HasSuffix(u.Host, ".visualstudio.com"):
		org := strings.TrimSuffix(u.Host, ".visualstudio.com")
		// Avoid doubling an existing subdomain like vssps.
		org = strings.SplitN(org, ".", 2)[0]
		u.Host = org + "." + subdomain + ".visualstudio.com"
	}
	return u.String(), nil
}
