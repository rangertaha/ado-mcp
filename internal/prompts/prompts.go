// SPDX-License-Identifier: MIT

// Package prompts registers MCP prompts: user-invoked, parameterized templates
// that clients surface as slash commands. Each prompt encodes a multi-step
// Azure DevOps workflow by guiding the model to call the right tools in order.
//
// Prompts are guidance only — they describe a procedure in terms of the tools
// exposed by this server. When the server runs read-only, prompts that suggest
// mutating steps simply won't find those tools available.
package prompts

import (
	"fmt"
	"strings"

	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the built-in workflow prompts to the server.
func Register(s *server.Server) {
	s.AddPrompt(
		"triage_bug",
		"Triage a work item: review it and propose (then apply) a state, assignee, and priority.",
		[]server.PromptArg{
			{Name: "project", Description: "project name or ID", Required: true},
			{Name: "id", Description: "work item ID", Required: true},
		},
		func(a map[string]string) string {
			return fmt.Sprintf(`Triage work item %s in project "%s".

Steps:
1. Call wit_get_work_item (project="%s", id=%s) to load its fields.
2. Summarize the title, current state, assignee, and description in 2-3 sentences.
3. Recommend: a target State (e.g. Active/Resolved/Closed), an assignee, and a Priority/Severity, with a one-line justification for each.
4. Ask me to confirm. On confirmation, call wit_update_work_item with the agreed System.State, System.AssignedTo, and Microsoft.VSTS.Common.Priority fields.
5. Optionally add a triage note with wit_add_comment.`,
				a["id"], a["project"], a["project"], a["id"])
		},
	)

	s.AddPrompt(
		"review_pr",
		"Review a pull request: summarize the change, surface risks, and offer to vote or comment.",
		[]server.PromptArg{
			{Name: "project", Description: "project name or ID", Required: true},
			{Name: "repo", Description: "repository name or ID", Required: true},
			{Name: "pullRequestId", Description: "pull request ID", Required: true},
		},
		func(a map[string]string) string {
			return fmt.Sprintf(`Review pull request %s in repo "%s" (project "%s").

Steps:
1. Call git_get_pull_request to load the PR (title, description, source/target branches, status).
2. Call git_list_pr_threads to see existing review discussion.
3. Inspect the changed files: use git_list_commits and git_get_file as needed to understand the diff on the source branch.
4. Produce a concise review: summary of intent, notable risks or bugs, and test/doc gaps.
5. Offer next actions and, on my confirmation: git_add_pr_comment to post feedback, and/or git_vote_pull_request to set my vote (10 approve, -10 reject).`,
				a["pullRequestId"], a["repo"], a["project"])
		},
	)

	s.AddPrompt(
		"sprint_status",
		"Summarize the current sprint for a team: iteration scope, board state, and risks.",
		[]server.PromptArg{
			{Name: "project", Description: "project name or ID", Required: true},
			{Name: "team", Description: "team name or ID", Required: true},
		},
		func(a map[string]string) string {
			return fmt.Sprintf(`Report the current sprint status for team "%s" in project "%s".

Steps:
1. Call boards_list_iterations (project="%s", team="%s") and identify the current iteration by date.
2. Run wit_query with a WIQL like: SELECT [System.Id],[System.Title],[System.State],[System.AssignedTo] FROM WorkItems WHERE [System.TeamProject]='%s' AND [System.IterationPath] UNDER '<current iteration path>' — then wit_get_work_items to expand the results.
3. Optionally call boards_list_boards and boards_get_board_columns to see column distribution.
4. Summarize: total items by state, items at risk (blocked/unassigned/aging), and a short outlook for sprint completion.`,
				a["team"], a["project"], a["project"], a["team"], a["project"])
		},
	)

	s.AddPrompt(
		"explore_repo",
		"Explore a Git repository: branches, recent activity, structure, and a README summary.",
		[]server.PromptArg{
			{Name: "project", Description: "project name or ID", Required: true},
			{Name: "repo", Description: "repository name or ID", Required: true},
		},
		func(a map[string]string) string {
			return fmt.Sprintf(`Give me an overview of the Git repository "%s" in project "%s".

Steps:
1. Call git_get_repository (project="%s", repo="%s") for the default branch and metadata.
2. Call git_list_refs with filter "heads/" to list branches, and git_list_commits (top=10) for recent activity.
3. Call git_get_file on "/README.md" (and "/" via git_get_file as needed) to understand the project; summarize what the repo is and its layout.
4. Report: purpose, primary language/structure, active branches, and recent change themes. Note any open pull requests via git_list_pull_requests (status=active).`,
				a["repo"], a["project"], a["project"], a["repo"])
		},
	)

	s.AddPrompt(
		"update_wiki_page",
		"Create or update a wiki page with content you provide or generate.",
		[]server.PromptArg{
			{Name: "project", Description: "project name or ID", Required: true},
			{Name: "wiki", Description: "wiki ID or name", Required: true},
			{Name: "path", Description: "page path, e.g. /Onboarding/Setup", Required: true},
			{Name: "topic", Description: "what the page should cover (optional)", Required: false},
		},
		func(a map[string]string) string {
			topic := a["topic"]
			step2 := "2. Ask me for the content (or I'll provide it), in Markdown."
			if topic != "" {
				step2 = fmt.Sprintf("2. Draft Markdown content for the page about: %s. Show it to me for approval.", topic)
			}
			return fmt.Sprintf(`Create or update the wiki page "%s" in wiki "%s" (project "%s").

Steps:
1. Call wiki_get_page (project="%s", wikiId="%s", path="%s", includeContent=true) to check whether the page already exists and see current content.
%s
3. On my confirmation, call wiki_create_or_update_page with the path and the final Markdown content. (Note: updating an existing page may require handling a version conflict.)
4. Confirm the result and show the page path.`,
				a["path"], a["wiki"], a["project"], a["project"], a["wiki"], a["path"], step2)
		},
	)

	s.AddPrompt(
		"process_work_log",
		"Turn my daily work-log entries into Azure DevOps tickets and 7pace time entries.",
		[]server.PromptArg{
			{Name: "date", Description: "work date YYYY-MM-DD to process (defaults to today)", Required: false},
			{Name: "project", Description: "Azure DevOps project for new tickets", Required: false},
		},
		func(a map[string]string) string {
			date := a["date"]
			if date == "" {
				date = "today"
			}
			proj := a["project"]
			if proj == "" {
				proj = "the entry's project (ask me if missing)"
			}
			return fmt.Sprintf(`Process my work-log entries for %s into Azure DevOps and 7pace.

Steps:
1. Call logs_list (date=%q if a real date, otherwise today; unlogged=true) to load entries not yet logged.
2. For each entry, summarize it and propose an action:
   - If it has no linked work item (workItemId is 0), offer to create a ticket in %s via wit_create_work_item (or macro_create_bug for bugs), using the entry's summary.
   - If it has minutes and hours are not yet logged, offer to record time via sevenpace_create_worklog against the linked work item (lengthSeconds = minutes*60, timestamp on the entry's date).
3. On my confirmation for each entry, perform the actions, then call logs_update to set workItemId, ticketCreated=true, and/or hoursLogged=true so it is not processed again.
4. Finish with a short summary of tickets created and hours logged.`,
				date, date, proj)
		},
	)

	s.AddPrompt(
		"log_my_day",
		"Log time in 7pace against the work items I worked on today.",
		[]server.PromptArg{
			{Name: "date", Description: "ISO date to log against, e.g. 2026-06-28 (defaults to today)", Required: false},
			{Name: "totalHours", Description: "total hours to distribute across the items (optional)", Required: false},
		},
		func(a map[string]string) string {
			date := a["date"]
			if date == "" {
				date = "today"
			}
			hours := a["totalHours"]
			distribute := "Ask me how long I spent on each."
			if hours != "" {
				distribute = fmt.Sprintf("Distribute %s hours across them, proposing a split for me to confirm.", hours)
			}
			return strings.Join([]string{
				fmt.Sprintf("Help me log time in 7pace for %s.", date),
				"",
				"Steps:",
				"1. Run wit_query with WIQL: SELECT [System.Id],[System.Title] FROM WorkItems WHERE [System.AssignedTo] = @Me AND [System.State] = 'Active' — then wit_get_work_items to show titles.",
				"2. " + distribute,
				"3. For each item, on my confirmation call sevenpace_create_worklog with the work item ID, the length in seconds, and a timestamp on the chosen date.",
				"4. Finish by listing what was logged via sevenpace_list_worklogs.",
			}, "\n")
		},
	)
}
