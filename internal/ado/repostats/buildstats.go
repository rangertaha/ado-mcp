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

type buildWire struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	Result     string `json:"result"`
	Definition struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"definition"`
}

// BuildDefinitionStat summarizes outcomes for one build definition.
type BuildDefinitionStat struct {
	Definition         string  `json:"definition" jsonschema:"build definition name"`
	Total              int     `json:"total" jsonschema:"total builds"`
	Succeeded          int     `json:"succeeded" jsonschema:"succeeded builds"`
	Failed             int     `json:"failed" jsonschema:"failed builds"`
	PartiallySucceeded int     `json:"partiallySucceeded" jsonschema:"partially succeeded builds"`
	Canceled           int     `json:"canceled" jsonschema:"canceled builds"`
	SuccessRate        float64 `json:"successRate" jsonschema:"succeeded / total, 0..1"`
}

// BuildStats summarizes build outcomes for a project.
type BuildStats struct {
	Scope         string                `json:"scope" jsonschema:"what was analyzed"`
	BuildsScanned int                   `json:"buildsScanned" jsonschema:"builds examined"`
	Truncated     bool                  `json:"truncated" jsonschema:"true if the scan limit was hit"`
	ByResult      map[string]int        `json:"byResult" jsonschema:"counts keyed by result"`
	Definitions   []BuildDefinitionStat `json:"definitions" jsonschema:"per-definition outcomes, busiest first"`
}

// BuildStatsForProject pages builds for a project (optionally a single
// definition) and tallies outcomes per definition.
func (s *service) BuildStatsForProject(ctx context.Context, project string, definitionID, maxBuilds int) (*BuildStats, error) {
	if maxBuilds <= 0 {
		maxBuilds = defaultMaxCommits
	}
	st := &BuildStats{Scope: "project:" + project, ByResult: map[string]int{}}
	defs := map[int]*BuildDefinitionStat{}

	continuation := ""
	truncated := false
	for st.BuildsScanned < maxBuilds {
		// Clamp the page size to the remaining budget so the scan honors
		// maxBuilds exactly instead of over-fetching a full page (e.g. tallying
		// 100 builds when maxBuilds is 10), matching scanRepo's commit paging.
		top := pageSize
		if remaining := maxBuilds - st.BuildsScanned; remaining < top {
			top = remaining
		}
		q := url.Values{}
		q.Set("$top", strconv.Itoa(top))
		if definitionID > 0 {
			q.Set("definitions", strconv.Itoa(definitionID))
		}
		if continuation != "" {
			q.Set("continuationToken", continuation)
		}
		path := fmt.Sprintf("/%s/_apis/build/builds", url.PathEscape(project))
		var page client.List[buildWire]
		resp, err := s.c.Org.Do(ctx, client.Request{Method: "GET", Path: path, Query: q, Out: &page})
		if err != nil {
			return nil, err
		}
		for _, b := range page.Value {
			st.BuildsScanned++
			result := b.Result
			if result == "" {
				result = b.Status // in-progress builds have no result yet
			}
			st.ByResult[result]++
			d := defs[b.Definition.ID]
			if d == nil {
				d = &BuildDefinitionStat{Definition: b.Definition.Name}
				defs[b.Definition.ID] = d
			}
			d.Total++
			switch b.Result {
			case "succeeded":
				d.Succeeded++
			case "failed":
				d.Failed++
			case "partiallySucceeded":
				d.PartiallySucceeded++
			case "canceled":
				d.Canceled++
			}
		}
		continuation = resp.ContinuationToken
		if continuation == "" || len(page.Value) == 0 {
			break
		}
		if st.BuildsScanned >= maxBuilds {
			truncated = true
			break
		}
	}

	out := make([]BuildDefinitionStat, 0, len(defs))
	for _, d := range defs {
		if d.Total > 0 {
			d.SuccessRate = float64(d.Succeeded) / float64(d.Total)
		}
		out = append(out, *d)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Total != out[j].Total {
			return out[i].Total > out[j].Total
		}
		return out[i].Definition < out[j].Definition
	})
	st.Definitions = out
	st.Truncated = truncated
	return st, nil
}
