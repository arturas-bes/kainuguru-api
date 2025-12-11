package resolvers

import (
	"context"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/graphql/model"
	"github.com/kainuguru/kainuguru-api/internal/middleware"
)

// Subscription resolvers - Phase 3.4 (Real-time features)
// These are stubs for now - will be implemented when WebSocket support is added

// ExpiredItemNotifications streams notifications for expired items
func (r *subscriptionResolver) ExpiredItemNotifications(ctx context.Context, shoppingListID string) (<-chan *model.ExpiredItemNotification, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	_ = userID
	_ = shoppingListID

	// TODO: Implement real-time subscription (Phase 3.4)
	// For now, return a channel that immediately closes to prevent blocking
	ch := make(chan *model.ExpiredItemNotification)
	close(ch)
	return ch, nil
}

// WizardSessionUpdates streams updates for an active wizard session
func (r *subscriptionResolver) WizardSessionUpdates(ctx context.Context, sessionID string) (<-chan *model.WizardSession, error) {
	// Require authentication
	userID, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("authentication required")
	}

	_ = userID
	_ = sessionID

	// TODO: Implement real-time subscription (Phase 3.4)
	// For now, return a channel that immediately closes to prevent blocking
	ch := make(chan *model.WizardSession)
	close(ch)
	return ch, nil
}
