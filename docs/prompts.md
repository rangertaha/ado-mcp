# Prompts (workflows)

MCP has no dedicated "workflow" primitive, so multi-step flows are shipped two ways.

**Prompts** are user-invoked templates that clients surface as slash commands. Each guides the model through a sequence of tool calls. Built-in prompts:

| Prompt | Arguments | What it does |
| ------ | --------- | ------------ |
| `triage_bug` | project, id | Load a work item, summarize it, propose & apply state/assignee/priority |
| `review_pr` | project, repo, pullRequestId | Summarize a PR and its threads, surface risks, offer to comment/vote |
| `sprint_status` | project, team | Current iteration scope, board state, items at risk |
| `explore_repo` | project, repo | Branches, recent commits, structure, README summary |
| `update_wiki_page` | project, wiki, path, topic | Draft/confirm content, then create or update a wiki page |
| `process_work_log` | date, project | Turn daily work-log journal entries into ADO tickets and 7pace time entries |
| `log_my_day` | date, totalHours | List my active work items and log 7pace time against them |

MCP clients surface these prompts as **slash commands** automatically (e.g. in Claude Code and Claude Desktop), so each is available as a slash command named after the prompt once the server is connected.

**Composite macro tools** (`macros` toolset) are single tools that orchestrate several REST calls server-side, for flows that are awkward one call at a time:

- `macro_complete_pull_request`: fetches the PR's last merge commit, then merges with the chosen options.
- `macro_create_bug`: creates a Bug from common fields and optionally adds a comment in one call.
- `macro_publish_wiki_page`: creates or updates a wiki page and reports which it did.
