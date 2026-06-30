// SPDX-License-Identifier: MIT

// Package repostats computes Azure DevOps surveys and statistics: repository
// contributions (commits per author), pull-request activity, work-item
// breakdowns, and build outcomes — aggregated at the repository, project, or
// organization level. It is exposed as the "stats" toolset.
//
// Azure DevOps has no first-class analytics endpoints for most of these, so the
// tools page the underlying REST APIs and aggregate in-process. Scans are
// bounded (see the max/top arguments) to stay responsive; when a bound is hit
// the result is flagged Truncated so callers know coverage was partial.
package repostats

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "stats"

// defaultMaxCommits bounds how many commits a single stats call will scan
// (per repository) when the caller does not specify a limit.
const defaultMaxCommits = 2000

// pageSize is the commits page size used when paging.
const pageSize = 100

// service wraps the Azure DevOps clients for contribution analytics.
type service struct {
	c *ado.Clients
}

// --- Wire types ---

type signature struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

type changeCounts struct {
	Add    int `json:"Add"`
	Edit   int `json:"Edit"`
	Delete int `json:"Delete"`
}

type commit struct {
	CommitID     string       `json:"commitId"`
	Author       signature    `json:"author"`
	ChangeCounts changeCounts `json:"changeCounts"`
}

type repoRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// --- Result types ---

// Contributor is a single author's aggregated contribution.
type Contributor struct {
	Name        string `json:"name" jsonschema:"author display name"`
	Email       string `json:"email" jsonschema:"author email (the aggregation key)"`
	Commits     int    `json:"commits" jsonschema:"number of commits authored"`
	Adds        int    `json:"adds" jsonschema:"files added across those commits (best effort)"`
	Edits       int    `json:"edits" jsonschema:"files edited (best effort)"`
	Deletes     int    `json:"deletes" jsonschema:"files deleted (best effort)"`
	FirstCommit string `json:"firstCommit,omitempty" jsonschema:"earliest commit date seen"`
	LastCommit  string `json:"lastCommit,omitempty" jsonschema:"latest commit date seen"`
}

// Stats is a contribution report for a scope (repo, project, or org).
type Stats struct {
	Scope          string        `json:"scope" jsonschema:"what was analyzed, e.g. repo:Foo or project:Bar or org"`
	Repositories   int           `json:"repositories" jsonschema:"number of repositories scanned"`
	SkippedRepos   int           `json:"skippedRepos" jsonschema:"repositories skipped because their commits could not be read"`
	CommitsScanned int           `json:"commitsScanned" jsonschema:"total commits examined"`
	Truncated      bool          `json:"truncated" jsonschema:"true if a scan limit was hit and coverage is partial"`
	Contributors   []Contributor `json:"contributors" jsonschema:"contributors sorted by commit count, descending"`
}

// aggregator accumulates per-author totals keyed by email (falling back to name).
type aggregator struct {
	byKey   map[string]*Contributor
	scanned int
}

func newAggregator() *aggregator { return &aggregator{byKey: map[string]*Contributor{}} }

func (a *aggregator) add(c commit) {
	key := c.Author.Email
	if key == "" {
		key = c.Author.Name
	}
	con := a.byKey[key]
	if con == nil {
		con = &Contributor{Name: c.Author.Name, Email: c.Author.Email}
		a.byKey[key] = con
	}
	con.Commits++
	con.Adds += c.ChangeCounts.Add
	con.Edits += c.ChangeCounts.Edit
	con.Deletes += c.ChangeCounts.Delete
	d := c.Author.Date
	if d != "" {
		if con.FirstCommit == "" || d < con.FirstCommit {
			con.FirstCommit = d
		}
		if d > con.LastCommit {
			con.LastCommit = d
		}
	}
	a.scanned++
}

// result builds the sorted contributor list.
func (a *aggregator) result(scope string, repos int, truncated bool) *Stats {
	out := make([]Contributor, 0, len(a.byKey))
	for _, c := range a.byKey {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Commits != out[j].Commits {
			return out[i].Commits > out[j].Commits
		}
		return out[i].Email < out[j].Email
	})
	return &Stats{
		Scope:          scope,
		Repositories:   repos,
		CommitsScanned: a.scanned,
		Truncated:      truncated,
		Contributors:   out,
	}
}

// scanRepo pages commits for one repository into the aggregator, up to max
// commits. It returns true if the limit was reached (i.e. results truncated).
func (s *service) scanRepo(ctx context.Context, agg *aggregator, project, repoID, fromDate, toDate string, max int) (bool, error) {
	skip := 0
	for {
		remaining := max - skip
		if remaining <= 0 {
			return true, nil
		}
		top := pageSize
		if remaining < top {
			top = remaining
		}

		q := url.Values{}
		q.Set("searchCriteria.$top", strconv.Itoa(top))
		q.Set("searchCriteria.$skip", strconv.Itoa(skip))
		if fromDate != "" {
			q.Set("searchCriteria.fromDate", fromDate)
		}
		if toDate != "" {
			q.Set("searchCriteria.toDate", toDate)
		}
		path := fmt.Sprintf("/%s/_apis/git/repositories/%s/commits", url.PathEscape(project), url.PathEscape(repoID))

		var page client.List[commit]
		if err := s.c.Org.GetJSON(ctx, path, q, &page); err != nil {
			return false, err
		}
		for _, c := range page.Value {
			agg.add(c)
		}
		got := len(page.Value)
		skip += got
		if got < top {
			return false, nil // exhausted this repo
		}
		if skip >= max {
			return true, nil // hit the cap
		}
	}
}

// listRepos returns the repositories in a project.
func (s *service) listRepos(ctx context.Context, project string) ([]repoRef, error) {
	var out client.List[repoRef]
	path := fmt.Sprintf("/%s/_apis/git/repositories", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// listProjects returns the projects in the organization.
func (s *service) listProjects(ctx context.Context) ([]projectRef, error) {
	var out client.List[projectRef]
	if err := s.c.Org.GetJSON(ctx, "/_apis/projects", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// RepoStats reports contributions for a single repository.
func (s *service) RepoStats(ctx context.Context, project, repo, fromDate, toDate string, maxCommits int) (*Stats, error) {
	if maxCommits <= 0 {
		maxCommits = defaultMaxCommits
	}
	agg := newAggregator()
	truncated, err := s.scanRepo(ctx, agg, project, repo, fromDate, toDate, maxCommits)
	if err != nil {
		return nil, err
	}
	return agg.result("repo:"+repo, 1, truncated), nil
}

// ProjectStats reports contributions aggregated across every repository in a
// project. maxCommits applies per repository.
func (s *service) ProjectStats(ctx context.Context, project, fromDate, toDate string, maxCommits int) (*Stats, error) {
	if maxCommits <= 0 {
		maxCommits = defaultMaxCommits
	}
	repos, err := s.listRepos(ctx, project)
	if err != nil {
		return nil, err
	}
	agg := newAggregator()
	truncated, skipped := false, 0
	for _, r := range repos {
		t, err := s.scanRepo(ctx, agg, project, r.ID, fromDate, toDate, maxCommits)
		if err != nil {
			skipped++ // e.g. a disabled or empty repo; skip rather than abort
			continue
		}
		truncated = truncated || t
	}
	st := agg.result("project:"+project, len(repos), truncated)
	st.SkippedRepos = skipped
	return st, nil
}

// OrgStats reports contributions aggregated across every repository in every
// project in the organization. maxCommits applies per repository; this can be
// expensive on large organizations.
func (s *service) OrgStats(ctx context.Context, fromDate, toDate string, maxCommits int) (*Stats, error) {
	if maxCommits <= 0 {
		maxCommits = defaultMaxCommits
	}
	projects, err := s.listProjects(ctx)
	if err != nil {
		return nil, err
	}
	agg := newAggregator()
	truncated := false
	repoCount, skipped := 0, 0
	for _, p := range projects {
		repos, err := s.listRepos(ctx, p.ID)
		if err != nil {
			continue // skip a project whose repos can't be listed
		}
		for _, r := range repos {
			repoCount++
			t, err := s.scanRepo(ctx, agg, p.ID, r.ID, fromDate, toDate, maxCommits)
			if err != nil {
				skipped++ // skip an unreadable repo rather than abort the whole org
				continue
			}
			truncated = truncated || t
		}
	}
	st := agg.result("org", repoCount, truncated)
	st.SkippedRepos = skipped
	return st, nil
}
