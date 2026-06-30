// SPDX-License-Identifier: MIT

package ado

import "testing"

func TestHostURL(t *testing.T) {
	cases := []struct {
		org, sub, want string
	}{
		{"https://dev.azure.com/myorg", "", "https://dev.azure.com/myorg"},
		{"https://dev.azure.com/myorg", "vssps", "https://vssps.dev.azure.com/myorg"},
		{"https://dev.azure.com/myorg", "vsrm", "https://vsrm.dev.azure.com/myorg"},
		{"https://dev.azure.com/myorg", "feeds", "https://feeds.dev.azure.com/myorg"},
		{"https://dev.azure.com/myorg", "auditservice", "https://auditservice.dev.azure.com/myorg"},
		{"https://myorg.visualstudio.com", "vssps", "https://myorg.vssps.visualstudio.com"},
		{"https://myorg.visualstudio.com", "vsrm", "https://myorg.vsrm.visualstudio.com"},
	}
	for _, c := range cases {
		got, err := hostURL(c.org, c.sub)
		if err != nil {
			t.Fatalf("hostURL(%q,%q): %v", c.org, c.sub, err)
		}
		if got != c.want {
			t.Errorf("hostURL(%q,%q) = %q, want %q", c.org, c.sub, got, c.want)
		}
	}
}

func TestNewClientsPopulatesAllHosts(t *testing.T) {
	c, err := NewClients("https://dev.azure.com/myorg", "pat")
	if err != nil {
		t.Fatal(err)
	}
	if c.Org == nil || c.VSSPS == nil || c.VSRM == nil || c.Feeds == nil || c.Audit == nil ||
		c.Search == nil || c.VSAEX == nil || c.ExtMgmt == nil || c.AdvSec == nil {
		t.Error("all host clients should be non-nil")
	}
}
