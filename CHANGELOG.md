# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **`ado` command-line tool** (built on `urfave/cli`): `ado mcp` runs the MCP
  server, `ado test` verifies Azure DevOps connectivity, and
  `ado log add|list|update|delete` manage the local work-log journal (no Azure
  DevOps credentials required for `log`).
- **Release tooling**: GoReleaser config producing archives for Linux, macOS,
  and Windows (amd64/arm64) plus a GitHub Release on tag push; `make snapshot`
  for local builds; svu-backed `make bump` / `make next` for
  conventional-commit versioning.
- **Documentation site**: MkDocs Material site published to GitHub Pages.
- **Version reporting** via `runtime/debug.ReadBuildInfo`, with an ldflags
  override (`internal.version`) for release builds.
- **Work-log journal** (`logs` toolset) — a local daily journal persisted to
  SQLite (via GORM) under the user's application directory
  (`<user config dir>/ado-mcp/data/ado-mcp.db`, overridable with `ADO_MCP_HOME`).
  Tools: `logs_add`, `logs_list`, `logs_get`, `logs_update`, `logs_delete`, `logs_summary`.
- **`process_work_log` prompt** — turns journal entries into Azure DevOps tickets
  and 7pace time entries, marking each entry as processed.
- **Application directory** — the server resolves and creates a per-user config
  directory on startup (`config.HomeDir`/`DataDir`).
- Client api-version negotiation (added in 0.1.x line) and broader unit tests
  across client, config, server, wiki, stats, wit, git, pipelines, release,
  boards, macros, 7pace, and the work-log store.

### Changed

- The binary is now `ado` and the MCP server is launched via the `ado mcp`
  subcommand (previously the standalone `ado-mcp` binary). Install path is now
  `github.com/rangertaha/ado-mcp/cmd/ado`. Update MCP client configs to run
  `ado mcp`.

## [0.1.0] - 2026-06-29

Initial release: a Go MCP server exposing broad Azure DevOps coverage plus
7pace Timetracker — **188 tools across 31 toolsets** and **6 workflow prompts**.

### Added

- **Server foundation**
  - MCP server built on the official `modelcontextprotocol/go-sdk`, served over stdio.
  - Generic JSON REST client: PAT (Basic) and bearer auth, default/per-request
    `api-version`, paging (continuation tokens, `$top`/`$skip`), raw-body support
    for logs, and structured `APIError` responses.
  - Automatic `api-version` negotiation: on a 400 "version not supported", the
    client reads the server's supported versions and retries once — transparently
    handling preview-only endpoints.
  - Typed tool registration with JSON Schema inferred from Go structs (input and
    output), MCP annotation hints (read-only / destructive / idempotent), and
    list results wrapped as objects.
  - Configuration via environment: `ADO_ORG_URL`, `ADO_PAT`, `SEVENPACE_ORG`,
    `SEVENPACE_TOKEN`, `ADO_TOOLSETS`, `ADO_READONLY`.
  - Multi-host routing: `dev.azure.com`, `vssps`, `vsrm`, `feeds`,
    `auditservice`, `almsearch`, `vsaex`, `extmgmt`, `advsec`.

- **Azure DevOps toolsets** — core, work items (incl. WIQL, comments, tags,
  classification nodes, queries, relations, attachments), git (repos, branches,
  commits, file commits, pull requests, threads, reviewers, votes, statuses),
  pipelines/build, release, test, boards, artifacts, wiki, graph, security,
  service hooks, dashboards, audit, distributed task (variable groups,
  environments, agent pools, secure files), service endpoints, policy, search
  (code/work-item/wiki), processes, member entitlement, notifications,
  extensions, TFVC, advanced security, approvals & checks, identities,
  operations, and profile.

- **Composite macros** — complete (merge) a pull request, create a bug with an
  optional comment, publish a wiki page (create/update), and publish a wiki page
  with image attachments in one call.

- **Statistics / surveys toolset** — contributor stats per repo/project/org,
  pull-request activity, work-item breakdowns, and build success rates.

- **7pace Timetracker** — current user, users, activity types, and worklog CRUD.

- **Workflow prompts** — `triage_bug`, `review_pr`, `sprint_status`,
  `explore_repo`, `update_wiki_page`, `log_my_day` — also shipped as Claude Code
  slash commands under `.claude/commands/`.

- **Project** — read-only mode, per-toolset filtering, README, MIT license,
  Makefile, and a GitHub Actions CI workflow (build/test/lint).

### Fixed

- Wiki page updates now send the required `If-Match` ETag, so editing an existing
  page works instead of failing with a precondition error (applied to both the
  `wiki` toolset and the `macro_publish_wiki_page` macro).
- Deleting a work-log entry that does not exist now reports "record not found"
  (and a non-zero exit) instead of falsely confirming the delete, matching the
  behavior of `log update` (`ado log delete` / `logs_delete`).
- Build-outcome stats now honor the `maxBuilds` cap exactly by clamping each
  page request to the remaining budget, instead of over-fetching a full page and
  tallying more builds than requested (`stats_builds`).
- Project-wide pull-request stats now skip a repository whose PRs cannot be read
  (reported via a new `skippedRepos` count) instead of aborting the entire scan,
  matching the commit-stats behavior (`stats_pull_requests`).
- API-version negotiation now compares against an `api-version` supplied via the
  request query and selects the newest matching supported version independent of
  the order the server lists them.

[Unreleased]: https://github.com/rangertaha/ado-mcp/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/rangertaha/ado-mcp/releases/tag/v0.1.0
