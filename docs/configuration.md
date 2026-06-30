# Configuration

All configuration is read from the environment.

| Variable           | Required | Description                                                                 |
| ------------------ | :------: | --------------------------------------------------------------------------- |
| `ADO_ORG_URL`      |   yes    | Organization URL, e.g. `https://dev.azure.com/myorg`.                        |
| `ADO_PAT`          |   yes    | Azure DevOps [Personal Access Token](https://learn.microsoft.com/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate). Sent via Basic auth. |
| `SEVENPACE_ORG`    |    no    | 7pace account name; builds `https://{org}.timehub.7pace.com/api`.            |
| `SEVENPACE_TOKEN`  |    no    | 7pace bearer token (*Settings, Reporting and API*). Enables 7pace tools.    |
| `ADO_TOOLSETS`     |    no    | Comma-separated toolset names to enable, or `all` (default).                 |
| `ADO_READONLY`     |    no    | `true` to expose only read-only tools.                                       |
| `ADO_MCP_HOME`     |    no    | Override the app directory (default: `<user config dir>/ado-mcp`, e.g. `~/.config/ado-mcp`). The work-log SQLite database lives in `data/` within it. |

7pace is optional: its tools register only when **both** `SEVENPACE_ORG` and `SEVENPACE_TOKEN` are set.

## Use with Claude Desktop / Claude Code

Add to your MCP client configuration (e.g. `claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "ado": {
      "command": "ado",
      "args": ["mcp"],
      "env": {
        "ADO_ORG_URL": "https://dev.azure.com/myorg",
        "ADO_PAT": "your-pat",
        "SEVENPACE_ORG": "myorg",
        "SEVENPACE_TOKEN": "your-7pace-token"
      }
    }
  }
}
```

For Claude Code: `claude mcp add ado --env ADO_ORG_URL=... --env ADO_PAT=... -- ado mcp`.

## Local development

The repo ships a committed [`.mcp.json`](https://github.com/rangertaha/ado-mcp/blob/main/.mcp.json) that runs the server straight from source (`go run ./cmd/ado mcp`), so changes take effect on the next session without a build step. It reads credentials from your environment (no secrets in the repo). Run `cp .env.example .env` and fill it in, or export `ADO_ORG_URL` and `ADO_PAT` (and optionally `SEVENPACE_ORG`/`SEVENPACE_TOKEN`) before launching Claude Code in this directory. The server is auto-trusted via `"enableAllProjectMcpServers": true` in the committed `.claude/settings.json`.
