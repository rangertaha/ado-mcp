# ado-mcp

[![CI](https://github.com/rangertaha/ado-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/rangertaha/ado-mcp/actions/workflows/ci.yml)

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server, written in Go, that exposes the **Azure DevOps** REST API (plus **7pace Timetracker**) as tools an LLM client (Claude Desktop/Code, Cursor, and others) can call.

It aims for broad coverage of Azure DevOps: **193 tools across 32 toolsets**, plus **7 workflow prompts** (also ready-made Claude Code slash commands), spanning projects, work items, Git/pull requests, pipelines, releases, test plans, boards, artifacts, wiki, graph/identity, security, service hooks, dashboards, audit, distributed task (variable groups/environments/agent pools), service connections, branch policies, code/work-item/wiki search, process customization, member entitlement, notifications, extensions, TFVC, advanced security alerts, approvals and checks, identities, operations, profile, composite macros, statistics/surveys, a local **work-log journal**, and 7pace time tracking.

A built-in workflow: keep a daily **work-log journal** (stored locally in SQLite under your config directory) and have Claude turn entries into Azure DevOps tickets and 7pace time entries. See the `logs` toolset and the `process_work_log` prompt.

## Features

- **Typed tools with schemas**: every tool has an auto-generated JSON Schema for its input and output, inferred from Go structs, with per-field descriptions. Inputs are validated before a handler runs.
- **Read-only switch**: `ADO_READONLY=true` hides every mutating tool, so the server can be safely pointed at production.
- **Toolset filtering**: enable only the areas you need with `ADO_TOOLSETS` to keep the tool list focused.
- **Multi-host aware**: transparently routes to the Azure DevOps service hosts (`dev.azure.com`, `vssps`, `vsrm`, `feeds`, `auditservice`, `almsearch`, `vsaex`, `extmgmt`, `advsec`).
- **Built on the official SDK**: uses [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) (v1).

## Next steps

- [Install](install.md) the server.
- Set up [Configuration](configuration.md) (environment variables and MCP client config).
- Browse the available [Toolsets](toolsets.md).
