// SPDX-License-Identifier: MIT

package servicehooks

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/server"
)

// Register wires the service hooks tools into the MCP server.
func Register(s *server.Server, c *ado.Clients) {
	s.NoteToolset(Name)
	svc := &service{c: c}

	server.Register(s, server.ToolDef{Name: "servicehooks_list_subscriptions", Title: "List subscriptions",
		Description: "List the service hook subscriptions in the organization."}, svc.listSubscriptions)
	server.Register(s, server.ToolDef{Name: "servicehooks_get_subscription", Title: "Get subscription",
		Description: "Get a single service hook subscription by ID."}, svc.getSubscription)
	server.Register(s, server.ToolDef{Name: "servicehooks_create_subscription", Title: "Create subscription",
		Description: "Create a new service hook subscription.", Write: true}, svc.createSubscription)
	server.Register(s, server.ToolDef{Name: "servicehooks_delete_subscription", Title: "Delete subscription",
		Description: "Delete a service hook subscription by ID.", Write: true, Destructive: true, Idempotent: true}, svc.deleteSubscription)
	server.Register(s, server.ToolDef{Name: "servicehooks_list_consumers", Title: "List consumers",
		Description: "List the available service hook consumers in the organization."}, svc.listConsumers)
	server.Register(s, server.ToolDef{Name: "servicehooks_list_publishers", Title: "List publishers",
		Description: "List the available service hook publishers in the organization."}, svc.listPublishers)
}

// EmptyInput is used by tools that take no arguments.
type EmptyInput struct{}

// GetSubscriptionInput identifies a subscription to retrieve.
type GetSubscriptionInput struct {
	SubscriptionID string `json:"subscriptionId" jsonschema:"the subscription ID"`
}

// CreateSubscriptionInput holds the free-form subscription definition to create.
type CreateSubscriptionInput struct {
	Subscription map[string]any `json:"subscription" jsonschema:"the free-form service hook subscription definition"`
}

// DeleteSubscriptionInput identifies a subscription to delete.
type DeleteSubscriptionInput struct {
	SubscriptionID string `json:"subscriptionId" jsonschema:"the subscription ID"`
}

func (s *service) listSubscriptions(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[Subscription], error) {
	out, err := s.ListSubscriptions(ctx)
	return nil, server.List(out), err
}

func (s *service) getSubscription(ctx context.Context, _ *mcp.CallToolRequest, in GetSubscriptionInput) (*mcp.CallToolResult, *Subscription, error) {
	out, err := s.GetSubscription(ctx, in.SubscriptionID)
	return nil, out, err
}

func (s *service) createSubscription(ctx context.Context, _ *mcp.CallToolRequest, in CreateSubscriptionInput) (*mcp.CallToolResult, *Subscription, error) {
	out, err := s.CreateSubscription(ctx, in.Subscription)
	return nil, out, err
}

func (s *service) deleteSubscription(ctx context.Context, _ *mcp.CallToolRequest, in DeleteSubscriptionInput) (*mcp.CallToolResult, *struct{}, error) {
	err := s.DeleteSubscription(ctx, in.SubscriptionID)
	return nil, &struct{}{}, err
}

func (s *service) listConsumers(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[Consumer], error) {
	out, err := s.ListConsumers(ctx)
	return nil, server.List(out), err
}

func (s *service) listPublishers(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, server.ListResult[Publisher], error) {
	out, err := s.ListPublishers(ctx)
	return nil, server.List(out), err
}
