// SPDX-License-Identifier: MIT

package wit

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register adds the Work Item Tracking toolset to the server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{
		Name:        "wit_get_work_item",
		Title:       "Get work item",
		Description: "Get a single work item by ID, including its fields.",
	}, svc.getWorkItem)

	server.Register(s, server.ToolDef{
		Name:        "wit_get_work_items",
		Title:       "Get work items (batch)",
		Description: "Get multiple work items by ID in one request. Optionally restrict to specific fields.",
	}, svc.getWorkItems)

	server.Register(s, server.ToolDef{
		Name:        "wit_query",
		Title:       "Run WIQL query",
		Description: "Run a Work Item Query Language (WIQL) query and return matching work item IDs. Example: \"SELECT [System.Id] FROM WorkItems WHERE [System.State] = 'Active'\".",
	}, svc.query)

	server.Register(s, server.ToolDef{
		Name:        "wit_list_comments",
		Title:       "List work item comments",
		Description: "List the comments on a work item.",
	}, svc.listComments)

	server.Register(s, server.ToolDef{
		Name:        "wit_list_work_item_types",
		Title:       "List work item types",
		Description: "List the work item types (e.g. Bug, User Story, Task) defined in a project.",
	}, svc.listWorkItemTypes)

	server.Register(s, server.ToolDef{
		Name:        "wit_list_fields",
		Title:       "List fields",
		Description: "List the work item field definitions in the organization.",
	}, svc.listFields)

	server.Register(s, server.ToolDef{
		Name:        "wit_list_tags",
		Title:       "List tags",
		Description: "List the work item tags defined in a project.",
	}, svc.listTags)

	server.Register(s, server.ToolDef{
		Name:        "wit_list_classification_nodes",
		Title:       "List classification nodes",
		Description: "List the classification node tree for a project. structureGroup is \"areas\" or \"iterations\". Optionally expand to a given depth.",
	}, svc.listClassificationNodes)

	server.Register(s, server.ToolDef{
		Name:        "wit_list_queries",
		Title:       "List queries",
		Description: "List the saved queries and query folders in a project. Optionally expand to a given depth.",
	}, svc.listQueries)

	// --- Write tools ---

	server.Register(s, server.ToolDef{
		Name:        "wit_create_work_item",
		Title:       "Create work item",
		Description: "Create a work item of a given type. Provide fields keyed by reference name, e.g. {\"System.Title\": \"Fix bug\", \"System.Description\": \"...\"}.",
		Write:       true,
	}, svc.createWorkItem)

	server.Register(s, server.ToolDef{
		Name:        "wit_update_work_item",
		Title:       "Update work item",
		Description: "Update fields on an existing work item. Provide fields keyed by reference name, e.g. {\"System.State\": \"Closed\"}.",
		Write:       true,
		Idempotent:  true,
	}, svc.updateWorkItem)

	server.Register(s, server.ToolDef{
		Name:        "wit_add_comment",
		Title:       "Add work item comment",
		Description: "Add a comment to a work item.",
		Write:       true,
	}, svc.addComment)

	server.Register(s, server.ToolDef{
		Name:        "wit_delete_work_item",
		Title:       "Delete work item",
		Description: "Delete a work item. By default it goes to the recycle bin; set destroy=true to permanently remove it.",
		Write:       true,
		Destructive: true,
		Idempotent:  true,
	}, svc.deleteWorkItem)

	server.Register(s, server.ToolDef{
		Name:        "wit_delete_tag",
		Title:       "Delete tag",
		Description: "Delete a work item tag by ID or name.",
		Write:       true,
		Destructive: true,
		Idempotent:  true,
	}, svc.deleteTag)

	server.Register(s, server.ToolDef{
		Name:        "wit_add_relation",
		Title:       "Add work item relation",
		Description: "Add a relation (link or attachment) to a work item. rel is the relation type, e.g. \"System.LinkTypes.Related\" or \"AttachedFile\".",
		Write:       true,
	}, svc.addRelation)

	server.Register(s, server.ToolDef{
		Name:        "wit_create_attachment",
		Title:       "Upload work item attachment",
		Description: "Upload an attachment and get a reference URL. Binary content must be base64-encoded. Then call wit_add_relation with rel=\"AttachedFile\" and the returned url to attach it to a work item.",
		Write:       true,
	}, svc.createAttachment)
}

// --- Tool input types ---

// GetWorkItemInput identifies a work item and optional expansion.
type GetWorkItemInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	ID      int    `json:"id" jsonschema:"work item ID"`
	Expand  string `json:"expand,omitempty" jsonschema:"optional expansion: none, relations, fields, links, or all"`
}

// GetWorkItemsInput is a batch fetch by IDs.
type GetWorkItemsInput struct {
	IDs    []int    `json:"ids" jsonschema:"work item IDs to fetch"`
	Fields []string `json:"fields,omitempty" jsonschema:"optional list of field reference names to return"`
}

// QueryInput runs a WIQL query in a project.
type QueryInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Wiql    string `json:"wiql" jsonschema:"the WIQL query text"`
	Top     int    `json:"top,omitempty" jsonschema:"maximum number of results (optional)"`
}

// WorkItemCommentsInput identifies a work item for comment listing.
type WorkItemCommentsInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	ID      int    `json:"id" jsonschema:"work item ID"`
}

// ProjectInput identifies a project.
type ProjectInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
}

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// CreateWorkItemInput describes a new work item.
type CreateWorkItemInput struct {
	Project string         `json:"project" jsonschema:"project name or ID"`
	Type    string         `json:"type" jsonschema:"work item type, e.g. Bug, Task, User Story"`
	Fields  map[string]any `json:"fields" jsonschema:"field values keyed by reference name, e.g. System.Title"`
}

// UpdateWorkItemInput describes field changes to an existing work item.
type UpdateWorkItemInput struct {
	ID     int            `json:"id" jsonschema:"work item ID"`
	Fields map[string]any `json:"fields" jsonschema:"field values keyed by reference name to set"`
}

// AddCommentInput adds a comment.
type AddCommentInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	ID      int    `json:"id" jsonschema:"work item ID"`
	Text    string `json:"text" jsonschema:"comment text (supports HTML)"`
}

// DeleteWorkItemInput identifies a work item to delete.
type DeleteWorkItemInput struct {
	ID      int  `json:"id" jsonschema:"work item ID"`
	Destroy bool `json:"destroy,omitempty" jsonschema:"permanently destroy instead of moving to recycle bin"`
}

// ListClassificationNodesInput identifies a classification node tree to list.
type ListClassificationNodesInput struct {
	Project        string `json:"project" jsonschema:"project name or ID"`
	StructureGroup string `json:"structureGroup" jsonschema:"node group: areas or iterations"`
	Depth          int    `json:"depth,omitempty" jsonschema:"optional depth to expand the tree"`
}

// ListQueriesInput identifies a project query tree to list.
type ListQueriesInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Depth   int    `json:"depth,omitempty" jsonschema:"optional depth to expand the tree"`
}

// DeleteTagInput identifies a tag to delete.
type DeleteTagInput struct {
	Project string `json:"project" jsonschema:"project name or ID"`
	Tag     string `json:"tag" jsonschema:"tag ID or name"`
}

// AddRelationInput describes a relation to add to a work item.
type AddRelationInput struct {
	ID      int    `json:"id" jsonschema:"work item ID"`
	Rel     string `json:"rel" jsonschema:"relation type, e.g. System.LinkTypes.Related or AttachedFile"`
	URL     string `json:"url" jsonschema:"URL of the related work item or attachment"`
	Comment string `json:"comment,omitempty" jsonschema:"optional comment describing the relation"`
}

// --- Tool handlers ---

func (s *service) getWorkItem(ctx context.Context, _ *mcp.CallToolRequest, in GetWorkItemInput) (*mcp.CallToolResult, *WorkItem, error) {
	out, err := s.GetWorkItem(ctx, in.Project, in.ID, in.Expand)
	return nil, out, err
}

func (s *service) getWorkItems(ctx context.Context, _ *mcp.CallToolRequest, in GetWorkItemsInput) (*mcp.CallToolResult, server.ListResult[WorkItem], error) {
	out, err := s.GetWorkItems(ctx, in.IDs, in.Fields)
	return nil, server.List(out), err
}

func (s *service) query(ctx context.Context, _ *mcp.CallToolRequest, in QueryInput) (*mcp.CallToolResult, *QueryResult, error) {
	out, err := s.Query(ctx, in.Project, in.Wiql, in.Top)
	return nil, out, err
}

func (s *service) listComments(ctx context.Context, _ *mcp.CallToolRequest, in WorkItemCommentsInput) (*mcp.CallToolResult, server.ListResult[Comment], error) {
	out, err := s.ListComments(ctx, in.Project, in.ID)
	return nil, server.List(out), err
}

func (s *service) listWorkItemTypes(ctx context.Context, _ *mcp.CallToolRequest, in ProjectInput) (*mcp.CallToolResult, server.ListResult[WorkItemType], error) {
	out, err := s.ListWorkItemTypes(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) listFields(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[Field], error) {
	out, err := s.ListFields(ctx)
	return nil, server.List(out), err
}

func (s *service) listTags(ctx context.Context, _ *mcp.CallToolRequest, in ProjectInput) (*mcp.CallToolResult, server.ListResult[Tag], error) {
	out, err := s.ListTags(ctx, in.Project)
	return nil, server.List(out), err
}

func (s *service) listClassificationNodes(ctx context.Context, _ *mcp.CallToolRequest, in ListClassificationNodesInput) (*mcp.CallToolResult, *Node, error) {
	out, err := s.ListClassificationNodes(ctx, in.Project, in.StructureGroup, in.Depth)
	return nil, out, err
}

func (s *service) listQueries(ctx context.Context, _ *mcp.CallToolRequest, in ListQueriesInput) (*mcp.CallToolResult, server.ListResult[Query], error) {
	out, err := s.ListQueries(ctx, in.Project, in.Depth)
	return nil, server.List(out), err
}

func (s *service) deleteTag(ctx context.Context, _ *mcp.CallToolRequest, in DeleteTagInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.DeleteTag(ctx, in.Project, in.Tag); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}

func (s *service) addRelation(ctx context.Context, _ *mcp.CallToolRequest, in AddRelationInput) (*mcp.CallToolResult, *WorkItem, error) {
	out, err := s.AddRelation(ctx, in.ID, in.Rel, in.URL, in.Comment)
	return nil, out, err
}

// CreateAttachmentInput uploads a work item attachment.
type CreateAttachmentInput struct {
	Project  string `json:"project" jsonschema:"project name or ID"`
	FileName string `json:"fileName" jsonschema:"attachment file name, e.g. log.txt"`
	Content  string `json:"content" jsonschema:"file content; binary data must be base64-encoded"`
}

func (s *service) createAttachment(ctx context.Context, _ *mcp.CallToolRequest, in CreateAttachmentInput) (*mcp.CallToolResult, *AttachmentReference, error) {
	out, err := s.CreateAttachment(ctx, in.Project, in.FileName, in.Content)
	return nil, out, err
}

func (s *service) createWorkItem(ctx context.Context, _ *mcp.CallToolRequest, in CreateWorkItemInput) (*mcp.CallToolResult, *WorkItem, error) {
	out, err := s.CreateWorkItem(ctx, in.Project, in.Type, in.Fields)
	return nil, out, err
}

func (s *service) updateWorkItem(ctx context.Context, _ *mcp.CallToolRequest, in UpdateWorkItemInput) (*mcp.CallToolResult, *WorkItem, error) {
	out, err := s.UpdateWorkItem(ctx, in.ID, in.Fields)
	return nil, out, err
}

func (s *service) addComment(ctx context.Context, _ *mcp.CallToolRequest, in AddCommentInput) (*mcp.CallToolResult, *Comment, error) {
	out, err := s.AddComment(ctx, in.Project, in.ID, in.Text)
	return nil, out, err
}

func (s *service) deleteWorkItem(ctx context.Context, _ *mcp.CallToolRequest, in DeleteWorkItemInput) (*mcp.CallToolResult, *struct{}, error) {
	if err := s.DeleteWorkItem(ctx, in.ID, in.Destroy); err != nil {
		return nil, nil, err
	}
	return nil, &struct{}{}, nil
}
