// SPDX-License-Identifier: MIT

package client

import (
	"encoding/base64"
	"net/http"
)

// PATAuthorizer authenticates to Azure DevOps using a Personal Access Token.
//
// Azure DevOps accepts a PAT via HTTP Basic auth with an empty username and the
// token as the password.
type PATAuthorizer struct {
	header string
}

// NewPATAuthorizer builds a PATAuthorizer for the given token.
func NewPATAuthorizer(pat string) *PATAuthorizer {
	encoded := base64.StdEncoding.EncodeToString([]byte(":" + pat))
	return &PATAuthorizer{header: "Basic " + encoded}
}

// Authorize sets the Authorization header for Basic PAT auth.
func (a *PATAuthorizer) Authorize(r *http.Request) {
	r.Header.Set("Authorization", a.header)
}

// BearerAuthorizer authenticates using an OAuth-style bearer token. It is used
// for 7pace Timetracker.
type BearerAuthorizer struct {
	header string
}

// NewBearerAuthorizer builds a BearerAuthorizer for the given token.
func NewBearerAuthorizer(token string) *BearerAuthorizer {
	return &BearerAuthorizer{header: "Bearer " + token}
}

// Authorize sets the Authorization header for bearer auth.
func (a *BearerAuthorizer) Authorize(r *http.Request) {
	r.Header.Set("Authorization", a.header)
}
