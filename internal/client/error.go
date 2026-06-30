// SPDX-License-Identifier: MIT

package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// APIError is a structured error returned for any non-2xx HTTP response. It
// preserves the HTTP status and, when available, the service-provided message
// so callers (and ultimately the LLM) get an actionable explanation rather than
// an opaque status code.
type APIError struct {
	Method     string // HTTP method of the failed request
	URL        string // request URL (path + query)
	StatusCode int    // HTTP status code
	Message    string // human-readable message extracted from the body
	TypeKey    string // Azure DevOps "typeKey", when present
	Body       string // raw response body (truncated)
}

// Error implements the error interface.
func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.Body
	}
	if msg == "" {
		msg = "(no response body)"
	}
	return fmt.Sprintf("%s %s -> HTTP %d: %s", e.Method, e.URL, e.StatusCode, msg)
}

// adoErrorBody mirrors the standard Azure DevOps error envelope.
type adoErrorBody struct {
	Message string `json:"message"`
	TypeKey string `json:"typeKey"`
}

// parseAPIError builds an *APIError from a failed response, best-effort decoding
// the Azure DevOps / 7pace error envelope.
func parseAPIError(method string, u *url.URL, status int, body []byte) *APIError {
	e := &APIError{
		Method:     method,
		URL:        u.RequestURI(),
		StatusCode: status,
		Body:       truncate(strings.TrimSpace(string(body)), 2000),
	}
	var env adoErrorBody
	if json.Unmarshal(body, &env) == nil {
		e.Message = env.Message
		e.TypeKey = env.TypeKey
	}
	return e
}

// truncate shortens s to at most n bytes, appending an ellipsis when cut.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// versionToken matches Azure DevOps api-version strings like "7.1" or
// "7.1-preview.3".
var versionToken = regexp.MustCompile(`\d+\.\d+(?:-preview(?:\.\d+)?)?`)

// negotiateAPIVersion inspects a failed response and returns a better
// api-version to retry with, or "" if no retry is warranted.
//
// Azure DevOps advertises the versions an endpoint accepts in the
// "Api-Supported-Versions" response header (and lists them in the error body).
// If the version we sent is NOT among them, this is a version mismatch — common
// causes are "not supported" and "under preview; the -preview flag must be
// supplied". We pick the supported version sharing our major.minor (e.g. we
// sent "7.1", the endpoint wants "7.1-preview.1"), falling back to the newest
// supported version. If the version we sent IS supported, the 400 is some other
// (input) error and we do not retry.
func negotiateAPIVersion(sent string, header http.Header, body []byte) string {
	if sent == "" {
		return ""
	}

	low := strings.ToLower(string(body))
	base := sent
	if i := strings.IndexByte(base, '-'); i >= 0 {
		base = base[:i] // "7.1-preview.2" -> "7.1"
	}
	supported := versionToken.FindAllString(header.Get("Api-Supported-Versions"), -1)

	// Case 1: the endpoint is preview-only and demands a "-preview" api-version
	// (Azure DevOps says "the -preview flag must be supplied", often with no
	// Api-Supported-Versions header). Prefer a supported preview sharing our
	// base; otherwise synthesize "<base>-preview".
	if strings.Contains(low, "preview") &&
		(strings.Contains(low, "-preview flag") || strings.Contains(low, "must be supplied") || strings.Contains(low, "under preview")) {
		best := ""
		for _, c := range supported {
			if sameBase(c, base) && strings.Contains(c, "preview") && (best == "" || moreRecent(c, best)) {
				best = c // pick the newest matching preview, not merely the last listed
			}
		}
		if best == "" {
			best = base + "-preview"
		}
		if best != sent {
			return best
		}
		return ""
	}

	// Case 2: an explicit supported list that does not include our version.
	if len(supported) == 0 {
		return ""
	}
	for _, c := range supported {
		if c == sent {
			return "" // our version is accepted; the 400 is some other error
		}
	}
	var prefixMatch, newest string
	for _, c := range supported {
		if newest == "" || moreRecent(c, newest) {
			newest = c
		}
		if sameBase(c, base) && (prefixMatch == "" || moreRecent(c, prefixMatch)) {
			prefixMatch = c // newest same-base version, independent of list order
		}
	}
	if prefixMatch != "" && prefixMatch != sent {
		return prefixMatch
	}
	if newest != sent {
		return newest
	}
	return ""
}

// sameBase reports whether candidate version c shares the same major.minor base
// (e.g. base "7.1" matches "7.1" and "7.1-preview.2" but not "7.10"). A plain
// prefix test would wrongly treat "7.10" as a match for "7.1".
func sameBase(c, base string) bool {
	if i := strings.IndexByte(c, '-'); i >= 0 {
		c = c[:i] // "7.1-preview.2" -> "7.1"
	}
	return c == base
}

// parseVersion splits an api-version like "7.1-preview.3" into comparable parts:
// major, minor, the preview ordinal (0 when absent), and whether it is a preview.
func parseVersion(v string) (major, minor, preview int, isPreview bool) {
	base := v
	if i := strings.IndexByte(v, '-'); i >= 0 {
		base = v[:i]
		rest := v[i+1:] // "preview" or "preview.3"
		isPreview = strings.HasPrefix(rest, "preview")
		if j := strings.IndexByte(rest, '.'); j >= 0 {
			preview, _ = strconv.Atoi(rest[j+1:])
		}
	}
	parts := strings.SplitN(base, ".", 2)
	major, _ = strconv.Atoi(parts[0])
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	return major, minor, preview, isPreview
}

// moreRecent reports whether api-version a is newer than b, without relying on
// the order in which the server lists supported versions. A stable release is
// considered newer than a preview of the same major.minor.
func moreRecent(a, b string) bool {
	amaj, amin, apre, aprev := parseVersion(a)
	bmaj, bmin, bpre, bprev := parseVersion(b)
	switch {
	case amaj != bmaj:
		return amaj > bmaj
	case amin != bmin:
		return amin > bmin
	case aprev != bprev:
		return !aprev // non-preview outranks preview at the same base
	default:
		return apre > bpre
	}
}
