// SPDX-License-Identifier: MIT

package repostats

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	"github.com/rangertaha/ado-mcp/internal/client"
)

// prWire is the subset of pull-request fields used for stats.
type prWire struct {
	PullRequestID int    `json:"pullRequestId"`
	Status        string `json:"status"`
	CreatedBy     struct {
		DisplayName string `json:"displayName"`
		UniqueName  string `json:"uniqueName"`
		ID          string `json:"id"`
	} `json:"createdBy"`
}

// PRAuthorStat is a single author's pull-request activity.
type PRAuthorStat struct {
	Name       string `json:"name" jsonschema:"author display name"`
	UniqueName string `json:"uniqueName,omitempty" jsonschema:"author unique name (email)"`
	Total      int    `json:"total" jsonschema:"pull requests created"`
	Active     int    `json:"active" jsonschema:"currently active"`
	Completed  int    `json:"completed" jsonschema:"completed (merged)"`
	Abandoned  int    `json:"abandoned" jsonschema:"abandoned"`
}

// PRStats is a pull-request activity report for a scope.
type PRStats struct {
	Scope        string         `json:"scope" jsonschema:"what was analyzed"`
	Repositories int            `json:"repositories" jsonschema:"repositories scanned"`
	PRsScanned   int            `json:"prsScanned" jsonschema:"pull requests examined"`
	Truncated    bool           `json:"truncated" jsonschema:"true if a scan limit was hit"`
	SkippedRepos int            `json:"skippedRepos" jsonschema:"repositories skipped because their pull requests could not be read"`
	ByStatus     map[string]int `json:"byStatus" jsonschema:"counts keyed by status"`
	Authors      []PRAuthorStat `json:"authors" jsonschema:"per-author activity, most active first"`
}

type prAggregator struct {
	byKey    map[string]*PRAuthorStat
	byStatus map[string]int
	scanned  int
}

func newPRAggregator() *prAggregator {
	return &prAggregator{byKey: map[string]*PRAuthorStat{}, byStatus: map[string]int{}}
}

func (a *prAggregator) add(pr prWire) {
	key := pr.CreatedBy.UniqueName
	if key == "" {
		key = pr.CreatedBy.DisplayName
	}
	s := a.byKey[key]
	if s == nil {
		s = &PRAuthorStat{Name: pr.CreatedBy.DisplayName, UniqueName: pr.CreatedBy.UniqueName}
		a.byKey[key] = s
	}
	s.Total++
	switch pr.Status {
	case "active":
		s.Active++
	case "completed":
		s.Completed++
	case "abandoned":
		s.Abandoned++
	}
	a.byStatus[pr.Status]++
	a.scanned++
}

func (a *prAggregator) result(scope string, repos int, truncated bool) *PRStats {
	out := make([]PRAuthorStat, 0, len(a.byKey))
	for _, s := range a.byKey {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].UniqueName < out[j].UniqueName
	})
	return &PRStats{Scope: scope, Repositories: repos, PRsScanned: a.scanned, Truncated: truncated, ByStatus: a.byStatus, Authors: out}
}

// scanPRs pages pull requests for a repository into the aggregator.
func (s *service) scanPRs(ctx context.Context, agg *prAggregator, project, repoID, status string, max int) (bool, error) {
	if status == "" {
		status = "all"
	}
	skip := 0
	for {
		if skip >= max {
			return true, nil
		}
		top := pageSize
		if max-skip < top {
			top = max - skip
		}
		q := url.Values{}
		q.Set("searchCriteria.status", status)
		q.Set("$top", strconv.Itoa(top))
		q.Set("$skip", strconv.Itoa(skip))
		path := fmt.Sprintf("/%s/_apis/git/repositories/%s/pullrequests", url.PathEscape(project), url.PathEscape(repoID))
		var page client.List[prWire]
		if err := s.c.Org.GetJSON(ctx, path, q, &page); err != nil {
			return false, err
		}
		for _, pr := range page.Value {
			agg.add(pr)
		}
		got := len(page.Value)
		skip += got
		if got < top {
			return false, nil
		}
	}
}

// PullRequestStats reports PR activity for a repo, or for all repos in a project
// when repo is empty.
func (s *service) PullRequestStats(ctx context.Context, project, repo, status string, maxPRs int) (*PRStats, error) {
	if maxPRs <= 0 {
		maxPRs = defaultMaxCommits
	}
	agg := newPRAggregator()
	if repo != "" {
		truncated, err := s.scanPRs(ctx, agg, project, repo, status, maxPRs)
		if err != nil {
			return nil, err
		}
		return agg.result("repo:"+repo, 1, truncated), nil
	}
	repos, err := s.listRepos(ctx, project)
	if err != nil {
		return nil, err
	}
	truncated, skipped := false, 0
	for _, r := range repos {
		t, err := s.scanPRs(ctx, agg, project, r.ID, status, maxPRs)
		if err != nil {
			skipped++ // e.g. a disabled or empty repo; skip rather than abort
			continue
		}
		truncated = truncated || t
	}
	out := agg.result("project:"+project, len(repos), truncated)
	out.SkippedRepos = skipped
	return out, nil
}
