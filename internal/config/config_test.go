// SPDX-License-Identifier: MIT

package config

import (
	"strings"
	"testing"
)

func TestLoad_RequiredFields(t *testing.T) {
	t.Setenv(EnvOrgURL, "")
	t.Setenv(EnvPAT, "")
	t.Setenv(EnvSevenPaceOrg, "")
	t.Setenv(EnvSevenPaceToken, "")
	t.Setenv(EnvToolsets, "")
	t.Setenv(EnvReadOnly, "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
	if !strings.Contains(err.Error(), EnvOrgURL) || !strings.Contains(err.Error(), EnvPAT) {
		t.Errorf("error should mention both required vars, got: %v", err)
	}
}

func TestLoad_Valid(t *testing.T) {
	t.Setenv(EnvOrgURL, "https://dev.azure.com/myorg/")
	t.Setenv(EnvPAT, "tok")
	t.Setenv(EnvSevenPaceOrg, "myorg")
	t.Setenv(EnvSevenPaceToken, "sptok")
	t.Setenv(EnvToolsets, "Core, WIT ,git")
	t.Setenv(EnvReadOnly, "yes")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OrgURL != "https://dev.azure.com/myorg" {
		t.Errorf("trailing slash not trimmed: %q", cfg.OrgURL)
	}
	if !cfg.ReadOnly {
		t.Error("ReadOnly should be true for 'yes'")
	}
	if !cfg.SevenPaceEnabled() {
		t.Error("7pace should be enabled")
	}
	if got := cfg.SevenPaceBaseURL(); got != "https://myorg.timehub.7pace.com/api/odata/v3.2" {
		t.Errorf("7pace base URL = %q", got)
	}
	if cfg.AllToolsets() {
		t.Error("AllToolsets should be false when a list is given")
	}
	if !cfg.ToolsetEnabled("git") || !cfg.ToolsetEnabled("WIT") {
		t.Error("listed toolsets should be enabled (case-insensitive)")
	}
	if cfg.ToolsetEnabled("release") {
		t.Error("unlisted toolset should be disabled")
	}
}

func TestLoad_PartialSevenPaceIsError(t *testing.T) {
	t.Setenv(EnvOrgURL, "https://dev.azure.com/o")
	t.Setenv(EnvPAT, "t")
	t.Setenv(EnvSevenPaceOrg, "o")
	t.Setenv(EnvSevenPaceToken, "")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for partial 7pace config")
	}
}

func TestToolsets_AllWhenEmpty(t *testing.T) {
	t.Setenv(EnvOrgURL, "https://dev.azure.com/o")
	t.Setenv(EnvPAT, "t")
	t.Setenv(EnvSevenPaceOrg, "")
	t.Setenv(EnvSevenPaceToken, "")
	t.Setenv(EnvToolsets, "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllToolsets() || !cfg.ToolsetEnabled("anything") {
		t.Error("empty toolset list should enable everything")
	}
	if cfg.SevenPaceEnabled() {
		t.Error("7pace should be disabled when unset")
	}
}
