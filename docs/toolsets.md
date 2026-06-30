# Toolsets

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
| `sevenpace`    | timehub.7pace  | worklogs, worklog↔work-item joins, work-item time rollups, budgets, raw OData query (read-only) |

See the full [Tools](tools.md) list for the individual tools in each toolset.
