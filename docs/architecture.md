# Architecture

```
cmd/ado                entrypoint: a urfave/cli command tree (mcp, test, log)
internal/config        environment configuration + validation
internal/client        generic JSON REST client (auth, api-version, paging, error mapping)
internal/server        MCP server wrapper, typed tool registration, read-only policy
internal/ado           Azure DevOps connection (one REST client per service host)
internal/ado/<area>    one package per service area: <area>.go (operations) + tools.go (tools)
internal/sevenpace     7pace Timetracker integration
```

The CLI exposes three commands: `mcp` (load config, register the enabled toolsets, and serve over stdio), `test` (verify credentials against Azure DevOps), and `log` (manage the local work-log journal).

Each service area follows the same shape: a `service` wrapping the shared REST clients exposes typed operations, and a `Register` function registers thin MCP tool handlers for them. Adding a new area is a matter of dropping in a package and listing it in `cmd/ado`.

## Development

```sh
make test        # go test -race ./...
make vet         # go vet ./...
make fmt-check   # gofmt verification
make all         # fmt-check + vet + test + build
```

## Smoke-testing the protocol

List the tools over stdio without an MCP client:

```sh
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"s","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
| ADO_ORG_URL=https://dev.azure.com/myorg ADO_PAT=$ADO_PAT ./bin/ado mcp
```

Or browse interactively with the [MCP Inspector](https://github.com/modelcontextprotocol/inspector):

```sh
npx @modelcontextprotocol/inspector ./bin/ado mcp
```
