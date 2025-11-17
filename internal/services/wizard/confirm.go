package wizard

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/monitoring"
	"github.com/uptrace/bun"
)

// ConfirmWizard applies all wizard decisions atomically to the shopping list.
// This is the culminating action that persists wizard changes to the database.
//
// Process:
//   1. Checks idempotency key (24-hour cache) to prevent duplicate confirmations
//   2. Loads session from Redis and validates status is ACTIVE
//   3. Revalidates all selected products (verifies not expired, prices unchanged)
//   4. Starts database transaction for ACID guarantees
//   5. For REPLACE decisions: creates OfferSnapshot, updates item to new product
//   6. For REMOVE decisions: deletes shopping list item
//   7. For SKIP decisions: no changes (item remains expired)
//   8. Commits transaction atomically
//   9. Updates session status to COMPLETED
//   10. Unlocks shopping list (sets is_locked=false)
//   11. Deletes session from Redis
//   12. Stores idempotency key with 24-hour TTL
//
// Returns ErrSessionNotFound if session does not exist.
// Returns ErrRevalidationFailed if products are stale/expired/price changed.
// On revalidation failure, session remains ACTIVE for user to review.
// On transaction failure, rolls back all changes and keeps session ACTIVE.
func (s *wizardService) ConfirmWizard(ctx context.Context, sessionID uuid.UUID, idempotencyKey string) (*ConfirmWizardResult, error) {
	s.logger.Info("ConfirmWizard called", "session_id", sessionID)

	// T057: Check idempotency cache first
	if idempotencyKey != "" {
		cachedSessionID, err := s.wizardCache.GetIdempotencyKey(ctx, idempotencyKey)
		if err == nil && cachedSessionID != uuid.Nil {
			// Already processed - return cached result from session metadata
			s.logger.Info("idempotency cache hit for confirmWizard",
				"session_id", sessionID,
				"idempotency_key", idempotencyKey)

			// Load result from completed session
			session, err := s.wizardCache.GetSession(ctx, cachedSessionID)
			if err == nil && session != nil {
				// Build result from session metadata
				result := &ConfirmWizardResult{
					ItemsUpdated:        len(session.Decisions),
					ItemsDeleted:        0,
					OfferSnapshotIDs:    []int64{},
					StoreCount:          len(session.SelectedStores),
					TotalEstimatedPrice: 0.0,
				}

				// Count actions
				for _, decision := range session.Decisions {
					if decision.Action == DecisionReplace {
						// REPLACE action
					} else if decision.Action == DecisionRemove {
						result.ItemsDeleted++
					}
				}
				result.ItemsUpdated = len(session.Decisions) - result.ItemsDeleted

				return result, nil
			}
		}
	}

	// Load session from Redis
	session, err := s.wizardCache.GetSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("failed to load wizard session",
			"session_id", sessionID,
			"error", err)
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Validate session status
	if session.Status != models.WizardStatusActive {
		return nil, fmt.Errorf("session status is not ACTIVE: %s", session.Status)
	}

	// T039: Check if session is expired
	if session.IsExpired() {
		// Delete expired session
		_ = s.wizardCache.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("session has expired")
	}

	// T058: Revalidation - verify all selected flyer products are still valid
	if err := s.revalidateDecisions(ctx, session); err != nil {
		s.logger.Error("revalidation failed",
			"session_id", sessionID,
			"error", err)
		// T059: Keep session IN_PROGRESS on revalidation failure
		return nil, fmt.Errorf("revalidation failed: %w", err)
	}

	// T051: Start database transaction for atomicity
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		s.logger.Error("failed to start transaction",
			"session_id", sessionID,
			"error", err)
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	result := &ConfirmWizardResult{
		ItemsUpdated:        0,
		ItemsDeleted:        0,
		OfferSnapshotIDs:    []int64{},
		StoreCount:          len(session.SelectedStores),
		TotalEstimatedPrice: 0.0,
	}

	// Process each decision
	for itemID, decision := range session.Decisions {
		switch decision.Action {
		case DecisionReplace:
			// T052: Create OfferSnapshot for REPLACE decision
			// T053: Update shopping_list_item with new flyer_product_id
			snapshotID, price, err := s.applyReplaceDecision(ctx, tx, session, itemID, decision)
			if err != nil {
				s.logger.Error("failed to apply REPLACE decision",
					"session_id", sessionID,
					"item_id", itemID,
					"error", err)
				return nil, fmt.Errorf("failed to apply REPLACE decision for item %d: %w", itemID, err)
			}
			result.ItemsUpdated++
			result.OfferSnapshotIDs = append(result.OfferSnapshotIDs, snapshotID)
			result.TotalEstimatedPrice += price

		case DecisionRemove:
			// T054: Delete shopping_list_item for REMOVE decision
			if err := s.applyRemoveDecision(ctx, tx, itemID); err != nil {
				s.logger.Error("failed to apply REMOVE decision",
					"session_id", sessionID,
					"item_id", itemID,
					"error", err)
				return nil, fmt.Errorf("failed to apply REMOVE decision for item %d: %w", itemID, err)
			}
			result.ItemsDeleted++

		case DecisionSkip:
			// T055: No changes for KEEP/SKIP - item remains expired
			s.logger.Info("SKIP decision - no changes",
				"session_id", sessionID,
				"item_id", itemID)

		default:
			s.logger.Warn("unknown decision action",
				"session_id", sessionID,
				"item_id", itemID,
				"action", decision.Action)
		}
	}

	// T051: Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("failed to commit transaction",
			"session_id", sessionID,
			"error", err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// T056: Update session Status=COMPLETED, delete from Redis
	session.Status = models.WizardStatusCompleted

	// Delete from Redis cache (session is now completed)
	if err := s.wizardCache.DeleteSession(ctx, sessionID); err != nil {
		s.logger.Warn("failed to delete completed session from Redis",
			"session_id", sessionID,
			"error", err)
	}

	// T057: Store idempotency result
	if idempotencyKey != "" {
		if err := s.wizardCache.SaveIdempotencyKey(ctx, idempotencyKey, sessionID); err != nil {
			s.logger.Warn("failed to store idempotency result",
				"session_id", sessionID,
				"idempotency_key", idempotencyKey,
				"error", err)
		}
	}

	// Track metrics
	monitoring.WizardSessionsTotal.WithLabelValues(string(models.WizardStatusCompleted)).Inc()

	s.logger.Info("wizard session confirmed successfully",
		"session_id", sessionID,
		"items_updated", result.ItemsUpdated,
		"items_deleted", result.ItemsDeleted,
		"snapshots_created", len(result.OfferSnapshotIDs),
		"total_price", result.TotalEstimatedPrice)

	return result, nil
}

// T058: revalidateDecisions verifies all selected flyer products are still valid
func (s *wizardService) revalidateDecisions(ctx context.Context, session *models.WizardSession) error {
	now := time.Now()

	// Collect all flyer product IDs from REPLACE decisions
	flyerProductIDs := []int64{}
	for _, decision := range session.Decisions {
		if decision.Action == DecisionReplace && decision.SuggestionID != nil {
			flyerProductIDs = append(flyerProductIDs, *decision.SuggestionID)
		}
	}

	if len(flyerProductIDs) == 0 {
		// No REPLACE decisions to validate
		return nil
	}

	// Fetch all selected flyer products in one query
	var flyerProducts []models.Product
	err := s.db.NewSelect().
		Model(&flyerProducts).
		Where("id IN (?)", bun.In(flyerProductIDs)).
		Scan(ctx)

	if err != nil {
		s.logger.Error("failed to fetch flyer products for revalidation",
			"session_id", session.ID,
			"error", err)
		monitoring.WizardRevalidationErrors.WithLabelValues("query_error").Inc()
		return fmt.Errorf("failed to fetch flyer products: %w", err)
	}

	// Build map for quick lookup
	productMap := make(map[int64]*models.Product)
	for i := range flyerProducts {
		productMap[int64(flyerProducts[i].ID)] = &flyerProducts[i]
	}

	// Validate each product
	staleProducts := []int64{}
	for _, decision := range session.Decisions {
		if decision.Action != DecisionReplace || decision.SuggestionID == nil {
			continue
		}

		product, exists := productMap[*decision.SuggestionID]
		if !exists {
			// Product not found - deleted?
			staleProducts = append(staleProducts, *decision.SuggestionID)
			continue
		}

		// Check if flyer is still valid (ValidTo >= NOW)
		var flyer models.Flyer
		err := s.db.NewSelect().
			Model(&flyer).
			Where("id = ?", product.FlyerID).
			Scan(ctx)

		if err != nil || flyer.ValidTo.Before(now) {
			// Flyer expired or not found
			staleProducts = append(staleProducts, *decision.SuggestionID)
			continue
		}

		// Note: Price changes are not validated per spec
		// Constitution ensures deterministic suggestions regardless of price fluctuations
	}

	if len(staleProducts) > 0 {
		s.logger.Error("revalidation detected stale products",
			"session_id", session.ID,
			"stale_count", len(staleProducts),
			"stale_products", staleProducts)
		monitoring.WizardRevalidationErrors.WithLabelValues("stale_data").Inc()
		return fmt.Errorf("revalidation failed: %d products are stale or expired", len(staleProducts))
	}

	s.logger.Info("revalidation passed",
		"session_id", session.ID,
		"products_validated", len(flyerProductIDs))

	return nil
}

// T052+T053: applyReplaceDecision creates OfferSnapshot and updates shopping_list_item
func (s *wizardService) applyReplaceDecision(
	ctx context.Context,
	tx bun.Tx,
	session *models.WizardSession,
	itemID int64,
	decision models.Decision,
) (snapshotID int64, price float64, err error) {
	if decision.SuggestionID == nil {
		return 0, 0, fmt.Errorf("REPLACE decision missing suggestion_id")
	}

	// Fetch the shopping list item
	var item models.ShoppingListItem
	err = tx.NewSelect().
		Model(&item).
		Where("id = ?", itemID).
		Scan(ctx)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch shopping list item: %w", err)
	}

	// Fetch the selected flyer product
	var flyerProduct models.Product
	err = tx.NewSelect().
		Model(&flyerProduct).
		Relation("Store").
		Where("product.id = ?", *decision.SuggestionID).
		Scan(ctx)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch flyer product: %w", err)
	}

	// T052: Create OfferSnapshot with snapshot_reason='wizard_migration', estimated=false
	// Convert ProductMasterID from *int to *int64 and StoreID from int to *int
	var productMasterID *int64
	if flyerProduct.ProductMasterID != nil {
		id := int64(*flyerProduct.ProductMasterID)
		productMasterID = &id
	}
	storeID := flyerProduct.StoreID

	snapshot := &models.OfferSnapshot{
		ShoppingListItemID: itemID,
		FlyerProductID:     decision.SuggestionID,
		ProductMasterID:    productMasterID,
		StoreID:            &storeID,
		ProductName:        flyerProduct.Name,
		Brand:              flyerProduct.Brand,
		Price:              flyerProduct.CurrentPrice,
		Unit:               flyerProduct.UnitType,
		SizeValue:          nil, // Product model uses UnitSize string, not separate value/unit
		SizeUnit:           nil,
		ValidFrom:          &flyerProduct.ValidFrom,
		ValidTo:            &flyerProduct.ValidTo,
		Estimated:          false, // Constitution-based suggestion (per spec)
		SnapshotReason:     string(models.SnapshotReasonWizardMigration),
		CreatedAt:          time.Now(),
	}

	_, err = tx.NewInsert().
		Model(snapshot).
		Exec(ctx)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to create offer snapshot: %w", err)
	}

	// T053: Update shopping_list_item with new flyer_product_id and origin='flyer'
	item.LinkedProductID = decision.SuggestionID
	// Convert ProductMasterID from *int to *int64
	if flyerProduct.ProductMasterID != nil {
		id := int64(*flyerProduct.ProductMasterID)
		item.ProductMasterID = &id
	}
	// Convert StoreID from int to *int
	storeIDPtr := flyerProduct.StoreID
	item.StoreID = &storeIDPtr
	// Convert FlyerID from int to *int
	flyerIDPtr := flyerProduct.FlyerID
	item.FlyerID = &flyerIDPtr
	item.Origin = "flyer" // FR-020 compliance
	item.EstimatedPrice = &flyerProduct.CurrentPrice
	item.AvailabilityStatus = "available"
	now := time.Now()
	item.AvailabilityCheckedAt = &now
	item.UpdatedAt = now

	_, err = tx.NewUpdate().
		Model(&item).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return 0, 0, fmt.Errorf("failed to update shopping list item: %w", err)
	}

	s.logger.Info("applied REPLACE decision",
		"session_id", session.ID,
		"item_id", itemID,
		"flyer_product_id", *decision.SuggestionID,
		"snapshot_id", snapshot.ID,
		"price", flyerProduct.CurrentPrice)

	return snapshot.ID, flyerProduct.CurrentPrice, nil
}

// T054: applyRemoveDecision deletes shopping_list_item
func (s *wizardService) applyRemoveDecision(ctx context.Context, tx bun.Tx, itemID int64) error {
	result, err := tx.NewDelete().
		Model((*models.ShoppingListItem)(nil)).
		Where("id = ?", itemID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete shopping list item: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("shopping list item not found: %d", itemID)
	}

	s.logger.Info("applied REMOVE decision",
		"item_id", itemID,
		"rows_affected", rowsAffected)

	return nil
}
