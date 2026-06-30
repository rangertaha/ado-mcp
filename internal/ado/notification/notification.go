// SPDX-License-Identifier: MIT

// Package notification exposes the Azure DevOps Notification service:
// listing and inspecting notification subscriptions and the available
// notification event types for the organization.
package notification

import (
	"context"
	"fmt"
	"net/url"

	"github.com/rangertaha/ado-mcp/internal/ado"
	"github.com/rangertaha/ado-mcp/internal/client"
)

// Name is the toolset name used for enable/disable filtering.
const Name = "notification"

// service holds the Azure DevOps host clients used by this area.
type service struct{ c *ado.Clients }

// --- Domain types ---

// NotificationSubscription represents a notification subscription that
// delivers events to a subscriber over a channel, subject to a filter.
type NotificationSubscription struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Subscriber  any    `json:"subscriber,omitempty"`
	Channel     any    `json:"channel,omitempty"`
	Filter      any    `json:"filter,omitempty"`
	Status      string `json:"status,omitempty"`
}

// EventType describes a notification event type that subscriptions can target.
type EventType struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Category    any    `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
}

// --- Notification operations ---

// ListSubscriptions returns the notification subscriptions for the organization.
func (s *service) ListSubscriptions(ctx context.Context) ([]NotificationSubscription, error) {
	var out client.List[NotificationSubscription]
	if err := s.c.Org.GetJSON(ctx, "/_apis/notification/subscriptions", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}

// GetSubscription returns a single notification subscription by its ID.
func (s *service) GetSubscription(ctx context.Context, subscriptionID string) (*NotificationSubscription, error) {
	var out NotificationSubscription
	path := fmt.Sprintf("/_apis/notification/subscriptions/%s", url.PathEscape(subscriptionID))
	if err := s.c.Org.GetJSON(ctx, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListEventTypes returns the available notification event types.
func (s *service) ListEventTypes(ctx context.Context) ([]EventType, error) {
	var out client.List[EventType]
	if err := s.c.Org.GetJSON(ctx, "/_apis/notification/eventtypes", nil, &out); err != nil {
		return nil, err
	}
	return out.Value, nil
}
