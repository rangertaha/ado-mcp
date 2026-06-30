// SPDX-License-Identifier: MIT

package notification

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the notification tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "notification_list_subscriptions", Title: "List subscriptions",
		Description: "List the notification subscriptions in the organization."}, svc.listSubscriptions)
	server.Register(s, server.ToolDef{Name: "notification_get_subscription", Title: "Get subscription",
		Description: "Get a single notification subscription by ID."}, svc.getSubscription)
	server.Register(s, server.ToolDef{Name: "notification_list_event_types", Title: "List event types",
		Description: "List the available notification event types in the organization."}, svc.listEventTypes)
}

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// GetSubscriptionInput identifies a notification subscription to retrieve.
type GetSubscriptionInput struct {
	SubscriptionID string `json:"subscriptionId" jsonschema:"the notification subscription ID"`
}

// --- Tool handlers ---

func (s *service) listSubscriptions(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[NotificationSubscription], error) {
	out, err := s.ListSubscriptions(ctx)
	return nil, server.List(out), err
}

func (s *service) getSubscription(ctx context.Context, _ *mcp.CallToolRequest, in GetSubscriptionInput) (*mcp.CallToolResult, *NotificationSubscription, error) {
	out, err := s.GetSubscription(ctx, in.SubscriptionID)
	return nil, out, err
}

func (s *service) listEventTypes(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[EventType], error) {
	out, err := s.ListEventTypes(ctx)
	return nil, server.List(out), err
}
