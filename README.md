# ado-mcp

[![CI](https://github.com/rangertaha/ado-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/rangertaha/ado-mcp/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/rangertaha/ado-mcp.svg)](https://pkg.go.dev/github.com/rangertaha/ado-mcp)
[![Go Report Card](https://goreportcard.com/badge/github.com/rangertaha/ado-mcp)](https://goreportcard.com/report/github.com/rangertaha/ado-mcp)
[![Go Version](https://img.shields.io/github/go-mod/go-version/rangertaha/ado-mcp)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server, written in Go, that exposes the **Azure DevOps** REST API (plus **7pace Timetracker**) as tools an LLM client (Claude Desktop/Code, Cursor, …) can call.

📖 **Documentation:** <https://rangertaha.github.io/ado-mcp/>

Broad coverage of Azure DevOps: **193 tools across 32 toolsets** plus **7 workflow prompts** (also ready-made Claude Code slash commands), spanning projects, work items, Git/pull requests, pipelines, releases, test plans, boards, artifacts, wiki, graph/identity, security, service hooks, dashboards, audit, distributed task (variable groups/environments/agent pools), service connections, branch policies, code/work-item/wiki search, process customization, member entitlement, notifications, extensions, TFVC, advanced security alerts, approvals & checks, identities, operations, profile, composite macros, statistics/surveys, a local **work-log journal**, and 7pace time tracking.

A built-in workflow: keep a daily **work-log journal** (stored locally in SQLite under your config directory) and have Claude turn entries into Azure DevOps tickets and 7pace time entries. See the `logs` toolset and the `process_work_log` prompt.

## Features

- **Typed tools with schemas**: every tool has an auto-generated JSON Schema for its input and output, inferred from Go structs, with per-field descriptions. Inputs are validated before a handler runs.
- **Read-only switch**: `ADO_READONLY=true` hides every mutating tool, so the server can be safely pointed at production.
- **Toolset filtering**: enable only the areas you need with `ADO_TOOLSETS` to keep the tool list focused.
- **Multi-host aware**: transparently routes to the Azure DevOps service hosts (`dev.azure.com`, `vssps`, `vsrm`, `feeds`, `auditservice`, `almsearch`, `vsaex`, `extmgmt`, `advsec`).
- **Built on the official SDK**: uses [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) (v1).

## Install

```sh
go install github.com/rangertaha/ado-mcp/cmd/ado@latest
```

This produces an `ado` binary. Or build from source:

```sh
git clone https://github.com/rangertaha/ado-mcp
cd ado-mcp
make build        # produces ./bin/ado
```

## CLI

`ado` is a small command tree (built on urfave/cli):

- `ado mcp`: run the MCP server over stdio. This is what MCP clients (Claude Desktop/Code, Cursor) invoke.
- `ado test`: verify connectivity and credentials against Azure DevOps (calls the profile endpoint and prints the authenticated user).
- `ado log`: manage the local SQLite work-log journal. These commands do not require Azure DevOps credentials.
  - `ado log add "<summary>" [--date YYYY-MM-DD] [--minutes|-m N] [--project|-p NAME] [--work-item|-w ID] [--activity|-a NAME] [--tags CSV]`: add an entry (summary is positional).
  - `ado log list [--date] [--from] [--to] [--limit N] [--unlogged]`: list entries, newest first.
  - `ado log update <id> [--summary|-s] [--date] [--minutes] [--project] [--work-item] [--activity] [--tags] [--ticket-created] [--hours-logged]`: update fields.
  - `ado log delete <id>` (alias `rm`): delete an entry.

## Configuration

All configuration is read from the environment.

| Variable           | Required | Description                                                                 |
| ------------------ | :------: | --------------------------------------------------------------------------- |
| `ADO_ORG_URL`      |   yes    | Organization URL, e.g. `https://dev.azure.com/myorg`.                        |
| `ADO_PAT`          |   yes    | Azure DevOps [Personal Access Token](https://learn.microsoft.com/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate). Sent via Basic auth. |
| `SEVENPACE_ORG`    |    no    | 7pace account name; builds `https://{org}.timehub.7pace.com/api`.            |
| `SEVENPACE_TOKEN`  |    no    | 7pace bearer token (*Settings → Reporting and API*). Enables 7pace tools.   |
| `ADO_TOOLSETS`     |    no    | Comma-separated toolset names to enable, or `all` (default).                 |
| `ADO_READONLY`     |    no    | `true` to expose only read-only tools.                                       |
| `ADO_MCP_HOME`     |    no    | Override the app directory (default: `<user config dir>/ado-mcp`, e.g. `~/.config/ado-mcp`). The work-log SQLite database lives in `data/` within it. |

7pace is optional: its tools register only when **both** `SEVENPACE_ORG` and `SEVENPACE_TOKEN` are set.

### Use with Claude Desktop / Claude Code

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

### Local development

The repo ships a committed [`.mcp.json`](.mcp.json) that runs the server straight from source (`go run ./cmd/ado mcp`), so changes take effect on the next session without a build step. It reads credentials from your environment (no secrets in the repo). Run `cp .env.example .env` and fill it in, or export `ADO_ORG_URL` and `ADO_PAT` (and optionally `SEVENPACE_ORG`/`SEVENPACE_TOKEN`) before launching Claude Code in this directory. The server is auto-trusted via `"enableAllProjectMcpServers": true` in the committed `.claude/settings.json`.

## Toolsets

Pass any subset to `ADO_TOOLSETS` (e.g. `ADO_TOOLSETS=core,wit,git`). 7pace uses the toolset name `sevenpace`.

| Toolset        | Host           | Covers                                                            |
| -------------- | -------------- | ---------------------------------------------------------------- |
| `core`         | dev.azure.com  | projects, teams, team members, processes                         |
| `wit`          | dev.azure.com  | work items (CRUD), WIQL, comments, types, fields, tags, attachments |
| `git`          | dev.azure.com  | repos, branches, commits, file commits, PRs, threads, reviewers, votes, statuses |
| `pipelines`    | dev.azure.com  | pipelines, runs, builds, logs, definitions, artifacts            |
| `release`      | vsrm           | release definitions, releases, deployments, approvals            |
| `test`         | dev.azure.com  | test plans, suites, cases, runs, results                         |
| `boards`       | dev.azure.com  | backlogs, iterations, capacity, boards, columns                  |
| `artifacts`    | feeds          | feeds, packages, package versions                                |
| `wiki`         | dev.azure.com  | wikis (CRUD/rename), pages (CRUD, append, list-tree, by-id, stats), attachments, moves |
| `graph`        | vssps          | users, groups, memberships                                       |
| `security`     | dev.azure.com  | security namespaces, access control lists                        |
| `servicehooks` | dev.azure.com  | subscriptions (CRUD)                                             |
| `dashboards`   | dev.azure.com  | dashboards, widgets                                              |
| `audit`        | auditservice   | audit log query                                                  |
| `distributedtask` | dev.azure.com | variable groups, environments, agent pools, secure files     |
| `serviceendpoint` | dev.azure.com | service connections (endpoints) and types                    |
| `policy`       | dev.azure.com  | branch/repo policy configurations and types                     |
| `search`       | almsearch      | code, work item, and wiki full-text search                      |
| `processes`    | dev.azure.com  | process customization: work item types, fields, states, behaviors |
| `memberentitlement` | vsaex     | user/group entitlements (licensing)                             |
| `notification` | dev.azure.com  | notification subscriptions and event types                      |
| `extension`    | extmgmt        | installed extensions                                             |
| `tfvc`         | dev.azure.com  | TFVC changesets, branches, items                                |
| `advsec`       | advsec         | Advanced Security code-scanning alerts                          |
| `approvals`    | dev.azure.com  | pipeline run approvals and checks                              |
| `identities`   | vssps          | legacy identity lookup                                          |
| `operations`   | dev.azure.com  | async operation status                                         |
| `profile`      | vssps          | the authenticated user's profile                              |
| `macros`       | dev.azure.com  | composite tools: complete PR, create bug, publish wiki page (+images) |
| `stats`        | dev.azure.com  | surveys: contributors (repo/project/org), PR/work-item/build stats |
| `logs`         | local (SQLite) | daily work-log journal: add/list/get/update/delete entries      |
| `sevenpace`    | timehub.7pace  | current user, users, activity types, worklogs (CRUD)            |

## Tools

Tools follow the naming convention `<toolset>_<verb>_<noun>`. Tools marked **[write]** mutate data and are hidden when `ADO_READONLY=true`.

<details>
<summary><strong>All 193 tools</strong></summary>

### core
- `core_get_project`: Get a single Azure DevOps project by name or ID.
- `core_get_team`: Get a single team within a project by name or ID.
- `core_list_processes`: List the process templates (e.g. Agile, Scrum, CMMI) available to the organization.
- `core_list_projects`: List the team projects in the Azure DevOps organization.
- `core_list_team_members`: List the members of a team within a project.
- `core_list_teams`: List the teams within a project.

### wit
- `wit_add_comment` **[write]**: Add a comment to a work item.
- `wit_add_relation` **[write]**: Add a relation (link or attachment) to a work item. rel is the relation type, e.g. "System.LinkTypes.Related" or "AttachedFile".
- `wit_create_attachment` **[write]**: Upload an attachment and get a reference URL. Binary content must be base64-encoded. Then call wit_add_relation with rel="AttachedFile" and the returned url to attach it to a work item.
- `wit_create_work_item` **[write]**: Create a work item of a given type. Provide fields keyed by reference name, e.g. {"System.Title": "Fix bug", "System.Description": "..."}.
- `wit_delete_tag` **[write]**: Delete a work item tag by ID or name.
- `wit_delete_work_item` **[write]**: Delete a work item. By default it goes to the recycle bin; set destroy=true to permanently remove it.
- `wit_get_work_item`: Get a single work item by ID, including its fields.
- `wit_get_work_items`: Get multiple work items by ID in one request. Optionally restrict to specific fields.
- `wit_list_classification_nodes`: List the classification node tree for a project. structureGroup is "areas" or "iterations". Optionally expand to a given depth.
- `wit_list_comments`: List the comments on a work item.
- `wit_list_fields`: List the work item field definitions in the organization.
- `wit_list_queries`: List the saved queries and query folders in a project. Optionally expand to a given depth.
- `wit_list_tags`: List the work item tags defined in a project.
- `wit_list_work_item_types`: List the work item types (e.g. Bug, User Story, Task) defined in a project.
- `wit_query`: Run a Work Item Query Language (WIQL) query and return matching work item IDs. Example: "SELECT [System.Id] FROM WorkItems WHERE [System.State] = 'Active'".
- `wit_update_work_item` **[write]**: Update fields on an existing work item. Provide fields keyed by reference name, e.g. {"System.State": "Closed"}.

### git
- `git_add_pr_comment` **[write]**: Add a new comment thread to a pull request.
- `git_add_pr_reviewer` **[write]**: Add (or update) a reviewer on a pull request, optionally marking them required.
- `git_create_branch` **[write]**: Create a new branch pointing at a commit SHA.
- `git_create_commit_status` **[write]**: Post a status against a commit (state: succeeded, failed, pending, error, notApplicable).
- `git_create_pull_request` **[write]**: Create a pull request from a source branch into a target branch.
- `git_create_repository` **[write]**: Create a new Git repository in a project.
- `git_delete_repository` **[write]**: Delete a Git repository by its ID.
- `git_get_file`: Get the text content of a file at a path, optionally at a specific branch.
- `git_get_pull_request`: Get a pull request by ID.
- `git_get_repository`: Get a Git repository by name or ID.
- `git_list_commit_statuses`: List the statuses posted against a commit.
- `git_list_commits`: List commits in a repository, optionally on a specific branch.
- `git_list_pr_reviewers`: List the reviewers on a pull request and their votes.
- `git_list_pr_threads`: List the comment threads on a pull request.
- `git_list_pull_requests`: List pull requests in a repository, filtered by status (active, completed, abandoned, all).
- `git_list_refs`: List refs in a repository. Use filter "heads/" for branches or "tags/" for tags.
- `git_list_repositories`: List the Git repositories in a project.
- `git_push_file` **[write]**: Add, edit, or delete a single file on a branch in one commit (push). changeType is add, edit, or delete; the branch tip is resolved automatically.
- `git_remove_pr_reviewer` **[write]**: Remove a reviewer from a pull request.
- `git_update_pull_request` **[write]**: Update a pull request: set status to completed or abandoned, or change title/description.
- `git_vote_pull_request` **[write]**: Set a reviewer's vote on a pull request (10 approve, 5 approve with suggestions, 0 reset, -5 waiting, -10 reject).

### pipelines
- `pipelines_cancel_build` **[write]**: Request cancellation of an in-progress build.
- `pipelines_get`: Get a pipeline by ID.
- `pipelines_get_build`: Get a single build by ID.
- `pipelines_get_build_changes`: Get the source changes associated with a build.
- `pipelines_get_build_log`: Get the text of a specific build log.
- `pipelines_get_build_timeline`: Get the execution timeline of a build.
- `pipelines_get_run`: Get a single pipeline run by ID.
- `pipelines_list`: List the pipelines in a project.
- `pipelines_list_artifacts`: List the artifacts produced by a build.
- `pipelines_list_build_logs`: List the logs produced by a build.
- `pipelines_list_builds`: List builds in a project, optionally filtered by build definition.
- `pipelines_list_definitions`: List the build definitions in a project.
- `pipelines_list_runs`: List the runs of a pipeline.
- `pipelines_run` **[write]**: Queue a new run of a pipeline, optionally on a specific branch.

### release
- `release_create_release` **[write]**: Create a new release from a release definition.
- `release_deploy_environment` **[write]**: Start a deployment of a release to an environment.
- `release_get_release`: Get a single release by ID.
- `release_list_approvals`: List release approvals in a project, optionally filtered by status.
- `release_list_definitions`: List the release definitions in a project.
- `release_list_deployments`: List deployments in a project, optionally filtered by release definition.
- `release_list_manual_interventions`: List the manual interventions for a release.
- `release_list_releases`: List releases in a project, optionally filtered by release definition.
- `release_update_approval` **[write]**: Approve or reject a release approval.
- `release_update_manual_intervention` **[write]**: Approve or reject a manual intervention in a release.

### test
- `test_add_results` **[write]**: Add test results to a test run.
- `test_create_run` **[write]**: Create a new manual test run for a test plan.
- `test_get_plan`: Get a single test plan by ID.
- `test_get_run`: Get a single test run by ID.
- `test_list_cases`: List the test cases in a test suite.
- `test_list_plans`: List the test plans in a project.
- `test_list_results`: List the results for a test run.
- `test_list_runs`: List the test runs in a project.
- `test_list_suites`: List the test suites in a test plan.
- `test_update_run` **[write]**: Update the state of a test run (e.g. InProgress, Completed, Aborted).

### boards
- `boards_get_board_columns`: Get the columns configured on a board.
- `boards_get_board_rows`: Get the rows (swimlanes) configured on a board.
- `boards_get_iteration_capacity`: Get the team capacity for a specific iteration.
- `boards_list_backlog_work_items`: List the work items contained in a backlog level.
- `boards_list_backlogs`: List the backlog levels configured for a team.
- `boards_list_board_charts`: List the charts available for a board.
- `boards_list_boards`: List the boards available to a team.
- `boards_list_iterations`: List the iterations (sprints) assigned to a team.

### artifacts
- `artifacts_get_feed`: Get a single packaging feed by its ID or name.
- `artifacts_get_package`: Get a single package within a feed by its ID.
- `artifacts_list_feed_views`: List the views defined on a packaging feed.
- `artifacts_list_feeds`: List the Azure Artifacts packaging feeds available to the caller.
- `artifacts_list_package_versions`: List the published versions of a package in a feed.
- `artifacts_list_packages`: List the packages contained in a feed.

### wiki
- `wiki_append_to_page` **[write]**: Append content to an existing wiki page (creating it if absent) without resending the whole page. Useful for adding changelog entries, notes, or sections.
- `wiki_create_attachment` **[write]**: Upload an attachment to a wiki. Binary content must be base64-encoded.
- `wiki_create_or_update_page` **[write]**: Create a new wiki page or update an existing one at the given path. Updating an existing page normally requires an If-Match ETag; this tool issues the PUT directly and the API may report a conflict if the page already exists.
- `wiki_create_wiki` **[write]**: Create a new project or code wiki in a project. type must be 'projectWiki' or 'codeWiki'.
- `wiki_delete_page` **[write]**: Delete a wiki page at the given path.
- `wiki_delete_wiki` **[write]**: Delete (unpublish) a wiki by ID or name.
- `wiki_get_page`: Get a wiki page by path, optionally including its content and (with recursionLevel='full') its subpages.
- `wiki_get_page_by_id`: Get a wiki page by its numeric page ID, optionally including content.
- `wiki_get_page_stats`: Get view statistics for a wiki page over the last N days.
- `wiki_get_wiki`: Get a single wiki by ID or name.
- `wiki_list_pages`: List the page tree of a wiki (expanded to full depth) from a root path, for browsing structure.
- `wiki_list_wikis`: List the wikis in a project.
- `wiki_move_page` **[write]**: Move or rename a wiki page from one path to another.
- `wiki_update_wiki` **[write]**: Rename a wiki.

### graph
- `graph_add_membership` **[write]**: Create a membership relating a subject to a container.
- `graph_get_descriptor`: Resolve a storage key (subject id) to its graph descriptor.
- `graph_list_groups`: List the group subjects in the organization graph.
- `graph_list_memberships`: List the memberships of a subject (optionally up to containers or down to members).
- `graph_list_users`: List the user subjects in the organization graph.
- `graph_remove_membership` **[write]**: Remove a membership relating a subject to a container.

### security
- `security_list_access_control_lists`: List the access control lists for a security namespace, optionally filtered by token.
- `security_list_namespaces`: List the security namespaces in the organization, optionally filtered by namespace ID.
- `security_remove_access_control_lists` **[write]**: Remove the access control lists for the given tokens in a security namespace.
- `security_set_access_control_entries` **[write]**: Add or update access control entries on the ACL for a token in a security namespace.

### servicehooks
- `servicehooks_create_subscription` **[write]**: Create a new service hook subscription.
- `servicehooks_delete_subscription` **[write]**: Delete a service hook subscription by ID.
- `servicehooks_get_subscription`: Get a single service hook subscription by ID.
- `servicehooks_list_consumers`: List the available service hook consumers in the organization.
- `servicehooks_list_publishers`: List the available service hook publishers in the organization.
- `servicehooks_list_subscriptions`: List the service hook subscriptions in the organization.

### dashboards
- `dashboards_create_dashboard` **[write]**: Create a new dashboard for a team within a project.
- `dashboards_delete_dashboard` **[write]**: Delete a dashboard by its ID for a team within a project.
- `dashboards_get_dashboard`: Get a single dashboard by its ID.
- `dashboards_list_dashboards`: List the dashboards for a team within a project.
- `dashboards_list_widgets`: List the widgets on a dashboard.

### audit
- `audit_query_log`: Query the organization-level Azure DevOps audit log. Returns a single page of decorated audit log entries; more pages may exist.

### dtask
- `dtask_create_environment` **[write]**: Create a deployment environment within a project.
- `dtask_create_variable_group` **[write]**: Create a variable group within a project from a raw body.
- `dtask_get_environment`: Get a single deployment environment by its ID.
- `dtask_get_variable_group`: Get a single variable group by its ID.
- `dtask_list_agent_pools`: List the organization-level agent pools.
- `dtask_list_environments`: List the deployment environments within a project.
- `dtask_list_secure_files`: List the secure files within a project.
- `dtask_list_variable_groups`: List the variable groups within a project.

### serviceendpoint
- `serviceendpoint_get`: Get a single service endpoint by ID within a project.
- `serviceendpoint_list`: List the service endpoints (service connections) in a project.
- `serviceendpoint_list_types`: List the available service endpoint types.

### policy
- `policy_create_configuration` **[write]**: Create a new policy configuration within a project.
- `policy_get_configuration`: Get a single policy configuration by its ID.
- `policy_list_configurations`: List the policy configurations for a project.
- `policy_list_types`: List the policy types available within a project.

### search
- `search_code`: Full-text search across source code in the organization, optionally scoped to a project.
- `search_wiki`: Full-text search across wiki pages in the organization, optionally scoped to a project.
- `search_work_items`: Full-text search across work items in the organization, optionally scoped to a project.

### processes
- `processes_get`: Get a single process by its process type ID.
- `processes_list`: List the work item tracking processes in the organization.
- `processes_list_behaviors`: List the behaviors defined within a process.
- `processes_list_fields`: List the fields of a work item type within a process.
- `processes_list_states`: List the workflow states of a work item type within a process.
- `processes_list_work_item_types`: List the work item types defined within a process.

### entitlement
- `entitlement_get_user`: Get the access-level entitlement for a single user by entitlement id.
- `entitlement_list_groups`: List group entitlements (licensing rules) for the organization.
- `entitlement_list_users`: List user access-level entitlements (licensing) for the organization.

### notification
- `notification_get_subscription`: Get a single notification subscription by ID.
- `notification_list_event_types`: List the available notification event types in the organization.
- `notification_list_subscriptions`: List the notification subscriptions in the organization.

### extension
- `extension_get`: Get a single installed extension by its publisher and extension identifiers.
- `extension_list_installed`: List all extensions installed in the Azure DevOps organization.

### tfvc
- `tfvc_get_changeset`: Get a single TFVC changeset by ID.
- `tfvc_get_item`: Get a TFVC item at the given path, including its content.
- `tfvc_list_branches`: List the TFVC branches, including child branches.
- `tfvc_list_changesets`: List TFVC changesets, optionally filtered by item path.

### advsec
- `advsec_get_alert`: Get a single Azure DevOps Advanced Security alert by its numeric ID.
- `advsec_list_alerts`: List Azure DevOps Advanced Security alerts for a repository, optionally filtered by alert type (code, secret or dependency).

### approvals
- `approvals_get`: Get a pipeline run approval by its identifier.
- `approvals_list_check_configurations`: List the check configurations defined on a protected resource (by resource type and id).
- `approvals_update` **[write]**: Approve or reject a pending pipeline run approval.

### identities
- `identities_read`: Look up legacy identities by search filter (e.g. 'General', 'DisplayName', 'AccountName'), filter value or descriptors.

### operations
- `operations_get`: Get the status of a long-running asynchronous Azure DevOps operation by its operation ID.

### profile
- `profile_get_me`: Retrieve the profile of the authenticated user.

### macro
- `macro_complete_pull_request` **[write]**: Complete (merge) a pull request. Fetches the PR's last merge commit automatically, then merges with the chosen options.
- `macro_create_bug` **[write]**: Create a Bug work item from common fields (title, repro steps, severity, assignee, area/iteration), optionally adding a comment in the same call.
- `macro_publish_wiki_page` **[write]**: Create or update a wiki page at a path with the given Markdown content, reporting whether it was created or updated.
- `macro_publish_wiki_page_with_images` **[write]**: Upload images as wiki attachments and create/update a page that embeds them. In the Markdown content, reference each image with the placeholder {{att:NAME}} (NAME matching an image's name); it is replaced with the attachment path. Image content must be base64-encoded.

### stats
- `stats_builds`: Build outcomes (success/failure rates) per definition for a project.
- `stats_org_contributors`: Contribution stats aggregated across every repository in every project in the organization. Can be slow on large orgs.
- `stats_project_contributors`: Contribution stats aggregated across every repository in a project.
- `stats_pull_requests`: Pull-request activity per author and by status, for a repo or for a whole project (omit repo).
- `stats_repo_contributors`: Contribution stats (commits and change counts per author) for one repository. Optionally filter by date range.
- `stats_work_items`: Work item breakdown by type, state, and assignee for a project (optionally scoped by a WIQL query).

### logs
- `logs_add` **[write]**: Record a daily work-log (journal) entry: what you worked on, optional minutes, project, linked work item, and 7pace activity. Date defaults to today.
- `logs_delete` **[write]**: Delete a work-log entry by ID.
- `logs_get`: Get a single work-log entry by ID.
- `logs_list`: List work-log entries, newest first. Filter by an exact date, a date range, or only entries whose hours have not yet been logged to 7pace.
- `logs_update` **[write]**: Update fields of a work-log entry, including marking that a ticket was created or hours were logged. Omitted fields are left unchanged.

### sevenpace
- `sevenpace_create_worklog` **[write]**: Create a 7pace worklog (time entry) against a work item.
- `sevenpace_delete_worklog` **[write]**: Delete a 7pace worklog by ID.
- `sevenpace_list_activity_types`: List the 7pace activity types (e.g. Development, Testing).
- `sevenpace_list_users`: List all 7pace Timetracker users.
- `sevenpace_list_worklogs`: List 7pace worklogs (time entries). Filter with an OData $filter expression such as "Timestamp ge 2026-01-01T00:00:00Z".
- `sevenpace_me`: Get the currently authenticated 7pace Timetracker user.
- `sevenpace_update_worklog` **[write]**: Update an existing 7pace worklog.

</details>

## Prompts (workflows)

MCP has no dedicated "workflow" primitive, so multi-step flows are shipped two ways:

**Prompts**: user-invoked templates that clients surface as slash commands. Each guides the model through a sequence of tool calls. Built-in prompts:

| Prompt | Arguments | What it does |
| ------ | --------- | ------------ |
| `triage_bug` | project, id | Load a work item, summarize it, propose & apply state/assignee/priority |
| `review_pr` | project, repo, pullRequestId | Summarize a PR and its threads, surface risks, offer to comment/vote |
| `sprint_status` | project, team | Current iteration scope, board state, items at risk |
| `explore_repo` | project, repo | Branches, recent commits, structure, README summary |
| `update_wiki_page` | project, wiki, path, topic | Draft/confirm content, then create or update a wiki page |
| `process_work_log` | date, project | Turn daily work-log journal entries into ADO tickets and 7pace time entries |
| `log_my_day` | date, totalHours | List my active work items and log 7pace time against them |

The same workflows ship as **Claude Code slash commands** in [`.claude/commands/`](.claude/commands) (`/ado-triage-bug`, `/ado-review-pr`, `/ado-sprint-status`, `/ado-explore-repo`, `/ado-update-wiki`, `/ado-log-day`); copy that directory into any repo that uses the server.

**Composite macro tools** (`macros` toolset): single tools that orchestrate several REST calls server-side, for flows that are awkward one call at a time:

- `macro_complete_pull_request`: fetches the PR's last merge commit, then merges with the chosen options.
- `macro_create_bug`: creates a Bug from common fields and optionally adds a comment in one call.
- `macro_publish_wiki_page`: creates or updates a wiki page and reports which it did.

## Architecture

```
cmd/ado                entrypoint: a urfave/cli command tree (mcp, test, log)
internal/config        environment configuration + validation
internal/client        generic JSON REST client (auth, api-version, paging, error mapping)
internal/server        MCP server wrapper, typed tool registration, read-only policy
internal/ado           Azure DevOps connection (one REST client per service host)
internal/ado/<area>    one package per service area: <area>.go (operations) + tools.go (tools)
internal/sevenpace     7pace Timetracker integration
```

The `mcp` subcommand loads config, registers the enabled toolsets, and serves over stdio; `test` checks credentials; `log` operates on the local work-log journal.

Each service area follows the same shape: a `service` wrapping the shared REST clients exposes typed operations, and a `Register` function registers thin MCP tool handlers for them. Adding a new area is a matter of dropping in a package and listing it in `cmd/ado`.

## Development

```sh
make test        # go test -race ./...
make vet         # go vet ./...
make fmt-check   # gofmt verification
make all         # fmt-check + vet + lint + test + build
```

### Smoke-testing the protocol

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

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

MIT. See [LICENSE](LICENSE).
