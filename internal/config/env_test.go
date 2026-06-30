// SPDX-License-Identifier: MIT

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	content := `# a comment
export ADO_ORG_URL=https://dev.azure.com/fromfile
ADO_PAT="quoted-token"
SEVENPACE_ORG='single'
BLANK=

# spaced
ADO_TOOLSETS = core,wit
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// An already-set non-empty var must win over the file.
	t.Setenv("ADO_ORG_URL", "https://dev.azure.com/fromenv")
	// These are empty so the file should populate them (t.Setenv restores them
	// after the test, unlike os.Unsetenv which would leak to other tests).
	t.Setenv("ADO_PAT", "")
	t.Setenv("SEVENPACE_ORG", "")
	t.Setenv("ADO_TOOLSETS", "")

	if err := LoadEnvFile(path); err != nil {
		t.Fatal(err)
	}

	if got := os.Getenv("ADO_ORG_URL"); got != "https://dev.azure.com/fromenv" {
		t.Errorf("existing env var should win, got %q", got)
	}
	if got := os.Getenv("ADO_PAT"); got != "quoted-token" {
		t.Errorf("double-quoted value = %q", got)
	}
	if got := os.Getenv("SEVENPACE_ORG"); got != "single" {
		t.Errorf("single-quoted value = %q", got)
	}
	if got := os.Getenv("ADO_TOOLSETS"); got != "core,wit" {
		t.Errorf("spaced key/value = %q", got)
	}
}

func TestLoadEnvFile_MissingIsNoError(t *testing.T) {
	if err := LoadEnvFile(filepath.Join(t.TempDir(), "nope.env")); err != nil {
		t.Errorf("missing file should not error, got %v", err)
	}
}
