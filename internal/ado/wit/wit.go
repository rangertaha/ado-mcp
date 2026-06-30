// SPDX-License-Identifier: MIT

// Package wit exposes the Azure DevOps Work Item Tracking service: work item
// CRUD, WIQL queries, comments, tags and metadata (types and fields).
package wit

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "wit"

// jsonPatchContentType is required by Azure DevOps for work item create/update.
const jsonPatchContentType = "application/json-patch+json"

// service wraps the Azure DevOps clients for Work Item Tracking operations.
type service struct {
	c *ado.Clients
}

// --- Domain types ---

// WorkItem is a work item. Fields holds the reference-name keyed field values
// (e.g. "System.Title", "System.State", "System.AssignedTo").
type WorkItem struct {
	ID     int            `json:"id"`
	Rev    int            `json:"rev,omitempty"`
	Fields map[string]any `json:"fields,omitempty"`
	URL    string         `json:"url,omitempty"`
}

// WorkItemRef is a lightweight reference returned by WIQL queries.
type WorkItemRef struct {
	ID  int    `json:"id"`
	URL string `json:"url,omitempty"`
}

// QueryResult is the result of a WIQL query.
type QueryResult struct {
	QueryType string        `json:"queryType,omitempty"`
	WorkItems []WorkItemRef `json:"workItems,omitempty"`
	AsOf      string        `json:"asOf,omitempty"`
	Columns   []FieldRef    `json:"columns,omitempty"`
}

// FieldRef references a field by name and reference name.
type FieldRef struct {
	ReferenceName string `json:"referenceName,omitempty"`
	Name          string `json:"name,omitempty"`
}

// Comment is a work item comment.
type Comment struct {
	ID          int    `json:"id,omitempty"`
	Text        string `json:"text,omitempty"`
	CreatedBy   any    `json:"createdBy,omitempty"`
	CreatedDate string `json:"createdDate,omitempty"`
}

// commentList is the comments endpoint envelope.
type commentList struct {
	TotalCount int       `json:"totalCount"`
	Comments   []Comment `json:"comments"`
}

// WorkItemType describes a work item type (e.g. Bug, User Story).
type WorkItemType struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        any    `json:"icon,omitempty"`
}

// Field describes a work item field definition.
type Field struct {
	ReferenceName string `json:"referenceName"`
	Name          string `json:"name"`
	Type          string `json:"type,omitempty"`
	ReadOnly      bool   `json:"readOnly,omitempty"`
}

// patchOp is a single JSON Patch operation for work item create/update.
type patchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// Tag is a work item tag.
type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Node is a classification node (area or iteration) in the project tree.
type Node struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	StructureType string `json:"structureType,omitempty"`
	Path          string `json:"path,omitempty"`
	HasChildren   bool   `json:"hasChildren,omitempty"`
	Children      any    `json:"children,omitempty"`
}

// Query describes a saved query or query folder.
type Query struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	IsFolder bool   `json:"isFolder,omitempty"`
	Wiql     string `json:"wiql,omitempty"`
	Children any    `json:"children,omitempty"`
}

// --- Operations ---

// GetWorkItem returns a single work item, optionally expanding relations.
func (s *service) GetWorkItem(ctx context.Context, project string, id int, expand string) (*WorkItem, error) {
	q := url.Values{}
	if expand != "" {
		q.Set("$expand", expand)
	}
	var wi WorkItem
	path := fmt.Sprintf("/%s/_apis/wit/workitems/%d", url.PathEscape(project), id)
	if err := s.c.Org.GetJSON(ctx, path, q, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

// GetWorkItems returns multiple work items by ID in a single batch request.
func (s *service) GetWorkItems(ctx context.Context, ids []int, fields []string) ([]WorkItem, error) {
	q := url.Values{}
	q.Set("ids", joinInts(ids))
	if len(fields) > 0 {
		q.Set("fields", strings.Join(fields, ","))
	} else {
		q.Set("$expand", "fields")
	}
	var out client.List[WorkItem]
	if err := s.c.Org.GetJSON(ctx, "/_apis/wit/workitems", q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// CreateWorkItem creates a work item of the given type with the supplied fields.
func (s *service) CreateWorkItem(ctx context.Context, project, witype string, fields map[string]any) (*WorkItem, error) {
	ops := make([]patchOp, 0, len(fields))
	for name, value := range fields {
		ops = append(ops, patchOp{Op: "add", Path: "/fields/" + name, Value: value})
	}
	var wi WorkItem
	path := fmt.Sprintf("/%s/_apis/wit/workitems/$%s", url.PathEscape(project), url.PathEscape(witype))
	if err := s.c.Org.PostJSONPatch(ctx, path, ops, &wi); err != nil {
		return nil, err
	}
	return &wi, nil
}

// UpdateWorkItem updates fields on an existing work item.
func (s *service) UpdateWorkItem(ctx context.Context, id int, fields map[string]any) (*WorkItem, error) {
	ops := make([]patchOp, 0, len(fields))
	for name, value := range fields {
		ops = append(ops, patchOp{Op: "add", Path: "/fields/" + name, Value: value})
	}
	var wi WorkItem
	path := fmt.Sprintf("/_apis/wit/workitems/%d", id)
	if err := s.c.Org.PatchJSON(ctx, path, nil, ops, &wi, jsonPatchContentType); err != nil {
		return nil, err
	}
	return &wi, nil
}

// DeleteWorkItem deletes (or destroys) a work item.
func (s *service) DeleteWorkItem(ctx context.Context, id int, destroy bool) error {
	q := url.Values{}
	if destroy {
		q.Set("destroy", "true")
	}
	path := fmt.Sprintf("/_apis/wit/workitems/%d", id)
	return s.c.Org.Delete(ctx, path, q, nil)
}

// Query runs a WIQL query and returns the matching work item references.
func (s *service) Query(ctx context.Context, project, wiql string, top int) (*QueryResult, error) {
	q := url.Values{}
	if top > 0 {
		q.Set("$top", strconv.Itoa(top))
	}
	body := map[string]string{"query": wiql}
	var res QueryResult
	path := fmt.Sprintf("/%s/_apis/wit/wiql", url.PathEscape(project))
	if err := s.c.Org.PostJSON(ctx, path, q, body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ListComments returns the comments on a work item.
func (s *service) ListComments(ctx context.Context, project string, id int) ([]Comment, error) {
	var out commentList
	path := fmt.Sprintf("/%s/_apis/wit/workItems/%d/comments", url.PathEscape(project), id)
	if err := s.c.Org.GetJSONVersion(ctx, path, nil, &out, "7.1-preview.4"); err != nil {
		return nil, err
	}
	return out.Comments, nil
}

// AddComment adds a comment to a work item.
func (s *service) AddComment(ctx context.Context, project string, id int, text string) (*Comment, error) {
	var c Comment
	path := fmt.Sprintf("/%s/_apis/wit/workItems/%d/comments", url.PathEscape(project), id)
	if err := s.c.Org.PostJSONVersion(ctx, path, map[string]string{"text": text}, &c, "7.1-preview.4"); err != nil {
		return nil, err
	}
	return &c, nil
}

// ListWorkItemTypes returns the work item types defined in a project.
func (s *service) ListWorkItemTypes(ctx context.Context, project string) ([]WorkItemType, error) {
	var out client.List[WorkItemType]
	path := fmt.Sprintf("/%s/_apis/wit/workitemtypes", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListFields returns the work item field definitions in the organization.
func (s *service) ListFields(ctx context.Context) ([]Field, error) {
	var out client.List[Field]
	if err := s.c.Org.GetJSON(ctx, "/_apis/wit/fields", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListTags returns the work item tags defined in a project.
func (s *service) ListTags(ctx context.Context, project string) ([]Tag, error) {
	var out client.List[Tag]
	path := fmt.Sprintf("/%s/_apis/wit/tags", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// DeleteTag deletes a work item tag by ID or name.
func (s *service) DeleteTag(ctx context.Context, project, tag string) error {
	path := fmt.Sprintf("/%s/_apis/wit/tags/%s", url.PathEscape(project), url.PathEscape(tag))
	return s.c.Org.Delete(ctx, path, nil, nil)
}

// ListClassificationNodes returns the classification node tree (areas or
// iterations) for a project, optionally expanded to the given depth.
func (s *service) ListClassificationNodes(ctx context.Context, project, structureGroup string, depth int) (*Node, error) {
	q := url.Values{}
	if depth > 0 {
		q.Set("$depth", strconv.Itoa(depth))
	}
	var n Node
	path := fmt.Sprintf("/%s/_apis/wit/classificationnodes/%s", url.PathEscape(project), url.PathEscape(structureGroup))
	if err := s.c.Org.GetJSON(ctx, path, q, &n); err != nil {
		return nil, err
	}
	return &n, nil
}

// ListQueries returns the saved query tree for a project, optionally expanded
// to the given depth.
func (s *service) ListQueries(ctx context.Context, project string, depth int) ([]Query, error) {
	q := url.Values{}
	if depth > 0 {
		q.Set("$depth", strconv.Itoa(depth))
	}
	var out client.List[Query]
	path := fmt.Sprintf("/%s/_apis/wit/queries", url.PathEscape(project))
	if err := s.c.Org.GetJSON(ctx, path, q, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// AddRelation adds a relation (link or attachment) to a work item.
func (s *service) AddRelation(ctx context.Context, id int, rel, relURL, comment string) (*WorkItem, error) {
	value := map[string]any{"rel": rel, "url": relURL}
	if comment != "" {
		value["attributes"] = map[string]any{"comment": comment}
	}
	ops := []patchOp{{Op: "add", Path: "/relations/-", Value: value}}
	var wi WorkItem
	path := fmt.Sprintf("/_apis/wit/workitems/%d", id)
	if err := s.c.Org.PatchJSON(ctx, path, nil, ops, &wi, jsonPatchContentType); err != nil {
		return nil, err
	}
	return &wi, nil
}

// AttachmentReference is returned after uploading a work item attachment; its
// URL can be linked to a work item with AddRelation (rel "AttachedFile").
type AttachmentReference struct {
	ID  string `json:"id,omitempty"`
	URL string `json:"url,omitempty"`
}

// CreateAttachment uploads attachment content and returns a reference. The
// content is sent verbatim; binary data should be base64-encoded by the caller.
// Use AddRelation with rel "AttachedFile" and the returned URL to attach it to
// a work item.
func (s *service) CreateAttachment(ctx context.Context, project, fileName, content string) (*AttachmentReference, error) {
	q := url.Values{}
	q.Set("fileName", fileName)
	path := fmt.Sprintf("/%s/_apis/wit/attachments", url.PathEscape(project))
	var ref AttachmentReference
	_, err := s.c.Org.Do(ctx, client.Request{
		Method:      "POST",
		Path:        path,
		Query:       q,
		Body:        strings.NewReader(content),
		ContentType: "application/octet-stream",
		Out:         &ref,
	})
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

// joinInts renders a slice of ints as a comma-separated string.
func joinInts(ids []int) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ",")
}
