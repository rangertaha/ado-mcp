// SPDX-License-Identifier: MIT

// Package app assembles the fully-configured ado-mcp server from configuration.
// It is shared by the command entry point (cmd/ado) and the integration
// tests (cmd/test), so the exact server the binary runs is the one under test.
package app

import (
	"log"
	"os"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/ado/advsec"
	"github.com/rangertaha/ado-mcp/internal/ado/approvals"
	"github.com/rangertaha/ado-mcp/internal/ado/artifacts"
	"github.com/rangertaha/ado-mcp/internal/ado/audit"
	"github.com/rangertaha/ado-mcp/internal/ado/boards"
	"github.com/rangertaha/ado-mcp/internal/ado/core"
	"github.com/rangertaha/ado-mcp/internal/ado/dashboards"
	"github.com/rangertaha/ado-mcp/internal/ado/distributedtask"
	"github.com/rangertaha/ado-mcp/internal/ado/extension"
	"github.com/rangertaha/ado-mcp/internal/ado/git"
	"github.com/rangertaha/ado-mcp/internal/ado/graph"
	"github.com/rangertaha/ado-mcp/internal/ado/identities"
	"github.com/rangertaha/ado-mcp/internal/ado/macros"
	"github.com/rangertaha/ado-mcp/internal/ado/memberentitlement"
	"github.com/rangertaha/ado-mcp/internal/ado/notification"
	"github.com/rangertaha/ado-mcp/internal/ado/operations"
	"github.com/rangertaha/ado-mcp/internal/ado/pipelines"
	"github.com/rangertaha/ado-mcp/internal/ado/policy"
	"github.com/rangertaha/ado-mcp/internal/ado/processes"
	"github.com/rangertaha/ado-mcp/internal/ado/profile"
	"github.com/rangertaha/ado-mcp/internal/ado/release"
	"github.com/rangertaha/ado-mcp/internal/ado/repostats"
	"github.com/rangertaha/ado-mcp/internal/ado/search"
	"github.com/rangertaha/ado-mcp/internal/ado/security"
	"github.com/rangertaha/ado-mcp/internal/ado/serviceendpoint"
	"github.com/rangertaha/ado-mcp/internal/ado/servicehooks"
	adotest "github.com/rangertaha/ado-mcp/internal/ado/test"
	"github.com/rangertaha/ado-mcp/internal/ado/tfvc"
	"github.com/rangertaha/ado-mcp/internal/ado/wiki"
	"github.com/rangertaha/ado-mcp/internal/ado/wit"
	"github.com/rangertaha/ado-mcp/internal/config"
	"github.com/rangertaha/ado-mcp/internal/logs"
	"github.com/rangertaha/ado-mcp/internal/prompts"
	"github.com/rangertaha/ado-mcp/internal/server"
	"github.com/rangertaha/ado-mcp/internal/sevenpace"
)

// Assemble builds the fully-configured server (all enabled toolsets, 7pace,
// the work-log journal, and prompts) and returns it with a cleanup function.
// version is reported to clients.
func Assemble(cfg *config.Config, version string) (*server.Server, func(), error) {
	clients, err := ado.NewClients(cfg.OrgURL, cfg.PAT)
	if err != nil {
		return nil, nil, err
	}

	srv := server.New("ado-mcp", version, cfg.ReadOnly)

	// Register every Azure DevOps toolset enabled by configuration.
	for _, ts := range toolsets() {
		if cfg.ToolsetEnabled(ts.Name) {
			ts.Register(srv, clients)
		}
	}

	// 7pace is registered only when configured and enabled.
	if cfg.SevenPaceEnabled() && cfg.ToolsetEnabled(sevenpace.Name) {
		sp, err := sevenpace.NewClient(cfg.SevenPaceBaseURL(), cfg.SevenPaceToken)
		if err != nil {
			return nil, nil, err
		}
		sevenpace.Register(srv, sp)
	}

	// Diagnostics go to stderr; stdout is reserved for the MCP protocol.
	log.SetOutput(os.Stderr)

	// Open the local work-log journal (SQLite, under the user's app data dir)
	// and register its toolset. Failure is non-fatal.
	cleanup := func() {}
	if cfg.ToolsetEnabled(logs.Name) {
		if store, err := OpenJournal(cfg); err != nil {
			log.Printf("ado-mcp: opening work-log journal failed: %v; journal disabled", err)
		} else {
			cleanup = func() { _ = store.Close() }
			logs.RegisterTools(srv, store)
		}
	}

	// Register the built-in workflow prompts.
	prompts.Register(srv)

	return srv, cleanup, nil
}

// OpenJournal ensures the data directory exists and opens the local work-log
// journal. It is shared by the assembled server and the `ado log` CLI commands
// so both open the journal the same way.
func OpenJournal(cfg *config.Config) (*logs.Store, error) {
	if err := cfg.EnsureDirs(); err != nil {
		return nil, err
	}
	return logs.Open(logs.DBPath(cfg.DataDir))
}

// toolsets returns every Azure DevOps toolset registrar, in registration order.
// New service areas are added here.
func toolsets() []server.Toolset {
	return []server.Toolset{
		{Name: core.Name, Register: core.Register},
		{Name: wit.Name, Register: wit.Register},
		{Name: git.Name, Register: git.Register},
		{Name: pipelines.Name, Register: pipelines.Register},
		{Name: release.Name, Register: release.Register},
		{Name: adotest.Name, Register: adotest.Register},
		{Name: boards.Name, Register: boards.Register},
		{Name: artifacts.Name, Register: artifacts.Register},
		{Name: wiki.Name, Register: wiki.Register},
		{Name: graph.Name, Register: graph.Register},
		{Name: security.Name, Register: security.Register},
		{Name: servicehooks.Name, Register: servicehooks.Register},
		{Name: dashboards.Name, Register: dashboards.Register},
		{Name: audit.Name, Register: audit.Register},
		{Name: distributedtask.Name, Register: distributedtask.Register},
		{Name: serviceendpoint.Name, Register: serviceendpoint.Register},
		{Name: policy.Name, Register: policy.Register},
		{Name: search.Name, Register: search.Register},
		{Name: processes.Name, Register: processes.Register},
		{Name: memberentitlement.Name, Register: memberentitlement.Register},
		{Name: notification.Name, Register: notification.Register},
		{Name: extension.Name, Register: extension.Register},
		{Name: tfvc.Name, Register: tfvc.Register},
		{Name: advsec.Name, Register: advsec.Register},
		{Name: approvals.Name, Register: approvals.Register},
		{Name: identities.Name, Register: identities.Register},
		{Name: operations.Name, Register: operations.Register},
		{Name: profile.Name, Register: profile.Register},
		{Name: macros.Name, Register: macros.Register},
		{Name: repostats.Name, Register: repostats.Register},
	}
}
