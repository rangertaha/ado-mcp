// SPDX-License-Identifier: MIT

// Package config loads and validates runtime configuration for the ado-mcp
// server from environment variables.
//
// All configuration is supplied via the environment so the server can run as a
// stdio subprocess launched by an MCP client (Claude Desktop/Code, Cursor, …),
// where command-line flags are awkward to pass.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// appName is the directory name created under the user's config directory.
const appName = "ado-mcp"

// Environment variable names recognised by the server.
const (
	EnvOrgURL         = "ADO_ORG_URL"        // e.g. https://dev.azure.com/myorg
	EnvPAT            = "ADO_PAT"            // Azure DevOps Personal Access Token
	EnvSevenPaceOrg   = "SEVENPACE_ORG"      // 7pace account name (builds the timehub URL)
	EnvSevenPaceToken = "SEVENPACE_TOKEN"    // 7pace bearer token
	EnvSevenPaceBase  = "SEVENPACE_API_BASE" // full 7pace API base URL override
	EnvToolsets       = "ADO_TOOLSETS"       // comma-separated toolset names, or "all"
	EnvReadOnly       = "ADO_READONLY"       // "true" disables all write tools
	EnvHome           = "ADO_MCP_HOME"       // overrides the application home directory
)

// sevenPaceODataVersion is the default 7pace OData API version used when no
// explicit SEVENPACE_API_BASE override is given.
const sevenPaceODataVersion = "v3.2"

// Config holds validated server configuration.
type Config struct {
	// OrgURL is the Azure DevOps organization base URL, e.g.
	// "https://dev.azure.com/myorg". It never has a trailing slash.
	OrgURL string

	// PAT is the Azure DevOps Personal Access Token used for Basic auth.
	PAT string

	// SevenPaceOrg is the 7pace account name used to build the API base URL
	// "https://{org}.timehub.7pace.com/api". Empty disables 7pace tools.
	SevenPaceOrg string

	// SevenPaceToken is the 7pace bearer token. Empty disables 7pace tools.
	SevenPaceToken string

	// SevenPaceBase, when set, is the full 7pace API base URL (e.g.
	// "https://myorg.timehub.7pace.com/api/odata/v3.2"). Overrides the URL
	// derived from SevenPaceOrg.
	SevenPaceBase string

	// Toolsets is the set of enabled toolset names. A nil/empty set means "all".
	Toolsets []string

	// ReadOnly, when true, causes write (mutating) tools to be skipped at
	// registration time so they are never exposed to the client.
	ReadOnly bool

	// HomeDir is the application's base directory under the user's config
	// directory (e.g. ~/.config/ado-mcp), or ADO_MCP_HOME when set.
	HomeDir string

	// DataDir is the data subdirectory within HomeDir (HomeDir/data).
	DataDir string
}

// resolveHome computes the application home directory: ADO_MCP_HOME if set,
// otherwise <user config dir>/ado-mcp (per-OS, via os.UserConfigDir).
func resolveHome() (string, error) {
	if h := strings.TrimSpace(os.Getenv(EnvHome)); h != "" {
		return h, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locating user config directory: %w", err)
	}
	return filepath.Join(base, appName), nil
}

// Dirs resolves the application's home and data directories without requiring
// Azure DevOps credentials. It is used by commands that only touch local state
// (such as the work-log journal), which must work even when no PAT is set.
func Dirs() (home, data string, err error) {
	home, err = resolveHome()
	if err != nil {
		return "", "", err
	}
	return home, filepath.Join(home, "data"), nil
}

// EnsureDirs creates the application's home and data directories if they do not
// already exist. It is safe to call repeatedly and is a no-op when the paths
// could not be resolved.
func (c *Config) EnsureDirs() error {
	if c.DataDir == "" {
		return nil
	}
	// 0o700: the data directory holds the user's private work-log database.
	if err := os.MkdirAll(c.DataDir, 0o700); err != nil {
		return fmt.Errorf("creating data directory %s: %w", c.DataDir, err)
	}
	// MkdirAll leaves an existing directory's mode untouched, so tighten it
	// explicitly: upgraders who created the dir under an older version (0o755)
	// should not keep a world-readable work-log database.
	if err := os.Chmod(c.DataDir, 0o700); err != nil {
		return fmt.Errorf("securing data directory %s: %w", c.DataDir, err)
	}
	return nil
}

// SevenPaceEnabled reports whether 7pace integration is configured. It needs a
// token plus either an account name or an explicit base URL.
func (c *Config) SevenPaceEnabled() bool {
	return c.SevenPaceToken != "" && (c.SevenPaceOrg != "" || c.SevenPaceBase != "")
}

// SevenPaceBaseURL returns the 7pace API base URL: the explicit
// SEVENPACE_API_BASE override if set, otherwise the OData base derived from the
// account name (https://{org}.timehub.7pace.com/api/odata/{version}).
func (c *Config) SevenPaceBaseURL() string {
	if c.SevenPaceBase != "" {
		return strings.TrimRight(c.SevenPaceBase, "/")
	}
	return fmt.Sprintf("https://%s.timehub.7pace.com/api/odata/%s", c.SevenPaceOrg, sevenPaceODataVersion)
}

// AllToolsets reports whether every toolset should be enabled.
func (c *Config) AllToolsets() bool {
	if len(c.Toolsets) == 0 {
		return true
	}
	for _, t := range c.Toolsets {
		if t == "all" {
			return true
		}
	}
	return false
}

// ToolsetEnabled reports whether the named toolset should be registered.
func (c *Config) ToolsetEnabled(name string) bool {
	if c.AllToolsets() {
		return true
	}
	for _, t := range c.Toolsets {
		if strings.EqualFold(t, name) {
			return true
		}
	}
	return false
}

// Load reads configuration from the process environment and validates it.
// It returns an error describing every problem found, so misconfiguration can
// be fixed in one pass.
func Load() (*Config, error) {
	cfg := &Config{
		OrgURL:         strings.TrimRight(strings.TrimSpace(os.Getenv(EnvOrgURL)), "/"),
		PAT:            strings.TrimSpace(os.Getenv(EnvPAT)),
		SevenPaceOrg:   strings.TrimSpace(os.Getenv(EnvSevenPaceOrg)),
		SevenPaceToken: strings.TrimSpace(os.Getenv(EnvSevenPaceToken)),
		SevenPaceBase:  strings.TrimSpace(os.Getenv(EnvSevenPaceBase)),
		Toolsets:       splitList(os.Getenv(EnvToolsets)),
		ReadOnly:       isTruthy(os.Getenv(EnvReadOnly)),
	}

	var errs []error
	if cfg.OrgURL == "" {
		errs = append(errs, fmt.Errorf("%s is required (e.g. https://dev.azure.com/myorg)", EnvOrgURL))
	} else if u, err := url.Parse(cfg.OrgURL); err != nil || u.Scheme == "" || u.Host == "" {
		errs = append(errs, fmt.Errorf("%s is not a valid URL: %q", EnvOrgURL, cfg.OrgURL))
	}
	if cfg.PAT == "" {
		errs = append(errs, fmt.Errorf("%s is required", EnvPAT))
	}
	// 7pace is optional, but partial configuration is almost certainly a mistake.
	hasLocator := cfg.SevenPaceOrg != "" || cfg.SevenPaceBase != ""
	if hasLocator != (cfg.SevenPaceToken != "") {
		errs = append(errs, fmt.Errorf("to enable 7pace set %s plus either %s or %s (or set none)", EnvSevenPaceToken, EnvSevenPaceOrg, EnvSevenPaceBase))
	}

	// Resolve (but do not yet create) the application directories. Failure to
	// locate them is non-fatal: the server still runs, just without local state.
	if home, data, err := Dirs(); err == nil {
		cfg.HomeDir = home
		cfg.DataDir = data
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return cfg, nil
}

// splitList parses a comma-separated environment value into a trimmed,
// lower-cased slice, dropping empty entries.
func splitList(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// isTruthy reports whether an environment value represents boolean true.
func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
