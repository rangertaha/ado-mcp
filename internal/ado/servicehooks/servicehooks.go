// SPDX-License-Identifier: MIT

// Package servicehooks provides access to Azure DevOps Service Hooks,
// allowing callers to list, inspect, create, and delete event subscriptions.
package servicehooks

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name for the service hooks area.
const Name = "servicehooks"

// service holds the Azure DevOps host clients used by this area.
type service struct{ c *ado.Clients }

// Subscription represents a service hook subscription that delivers events
// from a publisher to a consumer.
type Subscription struct {
	ID               string `json:"id"`
	EventType        string `json:"eventType,omitempty"`
	ConsumerID       string `json:"consumerId,omitempty"`
	ConsumerActionID string `json:"consumerActionId,omitempty"`
	Status           string `json:"status,omitempty"`
	PublisherID      string `json:"publisherId,omitempty"`
	URL              string `json:"url,omitempty"`
}

// ListSubscriptions returns the service hook subscriptions for the organization.
func (s *service) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	var out client.List[Subscription]
	if err := s.c.Org.GetJSON(ctx, "/_apis/hooks/subscriptions", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetSubscription returns a single service hook subscription by its ID.
func (s *service) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	var out Subscription
	path := fmt.Sprintf("/_apis/hooks/subscriptions/%s", url.PathEscape(subscriptionID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateSubscription creates a new service hook subscription from a free-form
// subscription definition and returns the created subscription.
func (s *service) CreateSubscription(ctx context.Context, subscription map[string]any) (*Subscription, error) {
	var out Subscription
	if err := s.c.Org.PostJSON(ctx, "/_apis/hooks/subscriptions", nil, subscription, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSubscription deletes a service hook subscription by its ID.
func (s *service) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	path := fmt.Sprintf("/_apis/hooks/subscriptions/%s", url.PathEscape(subscriptionID))
	return s.c.Org.Delete(ctx, path, nil, nil)
}

// Consumer represents a service hook consumer that can receive events.
type Consumer struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// Publisher represents a service hook publisher that emits events.
type Publisher struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// ListConsumers returns the available service hook consumers.
func (s *service) ListConsumers(ctx context.Context) ([]Consumer, error) {
	var out client.List[Consumer]
	if err := s.c.Org.GetJSON(ctx, "/_apis/hooks/consumers", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// ListPublishers returns the available service hook publishers.
func (s *service) ListPublishers(ctx context.Context) ([]Publisher, error) {
	var out client.List[Publisher]
	if err := s.c.Org.GetJSON(ctx, "/_apis/hooks/publishers", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
