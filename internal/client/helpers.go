// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"net/http"
	"net/url"
)

// List is the standard Azure DevOps collection envelope: most list endpoints
// return {"count": N, "value": [...]}. Decode into List[T] to access the items.
type List[T any] struct {
	Count int `json:"count"`
	Value []T `json:"value"`
}

// GetJSON performs a GET request and decodes the response into out.
func (c *Client) GetJSON(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodGet, Path: path, Query: query, Out: out})
	return err
}

// PostJSON performs a POST request with a JSON body and decodes the response.
func (c *Client) PostJSON(ctx context.Context, path string, query url.Values, body, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPost, Path: path, Query: query, Body: body, Out: out})
	return err
}

// PutJSON performs a PUT request with a JSON body and decodes the response.
func (c *Client) PutJSON(ctx context.Context, path string, query url.Values, body, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPut, Path: path, Query: query, Body: body, Out: out})
	return err
}

// PatchJSON performs a PATCH request with a JSON body and decodes the response.
// contentType may be empty to use "application/json"; Azure DevOps work-item
// updates require "application/json-patch+json".
func (c *Client) PatchJSON(ctx context.Context, path string, query url.Values, body, out any, contentType string) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPatch, Path: path, Query: query, Body: body, Out: out, ContentType: contentType})
	return err
}

// Delete performs a DELETE request, optionally decoding a response body.
func (c *Client) Delete(ctx context.Context, path string, query url.Values, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodDelete, Path: path, Query: query, Out: out})
	return err
}

// GetJSONVersion is like GetJSON but pins a specific api-version (e.g. for
// "-preview" endpoints).
func (c *Client) GetJSONVersion(ctx context.Context, path string, query url.Values, out any, apiVersion string) error {
	_, err := c.Do(ctx, Request{Method: http.MethodGet, Path: path, Query: query, Out: out, APIVersion: apiVersion})
	return err
}

// PostJSONVersion is like PostJSON but pins a specific api-version.
func (c *Client) PostJSONVersion(ctx context.Context, path string, body, out any, apiVersion string) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPost, Path: path, Body: body, Out: out, APIVersion: apiVersion})
	return err
}

// PostJSONPatch performs a POST with the "application/json-patch+json" content
// type required by Azure DevOps work item create operations.
func (c *Client) PostJSONPatch(ctx context.Context, path string, body, out any) error {
	_, err := c.Do(ctx, Request{Method: http.MethodPost, Path: path, Body: body, Out: out, ContentType: "application/json-patch+json"})
	return err
}
