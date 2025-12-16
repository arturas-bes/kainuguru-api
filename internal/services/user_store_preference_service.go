package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/uptrace/bun"
)

// UserStorePreferenceService defines the interface for user store preference operations
type UserStorePreferenceService interface {
	// GetPreferredStoreIDs returns the store IDs that a user has marked as preferred
	GetPreferredStoreIDs(ctx context.Context, userID uuid.UUID) ([]int, error)

	// GetPreferredStores returns the full store objects for a user's preferred stores
	GetPreferredStores(ctx context.Context, userID uuid.UUID) ([]*models.Store, error)

	// SetPreferredStores sets all preferred stores for a user (replaces existing)
	SetPreferredStores(ctx context.Context, userID uuid.UUID, storeIDs []int) error

	// AddPreferredStore adds a single store to user's preferences
	AddPreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error

	// RemovePreferredStore removes a single store from user's preferences
	RemovePreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error

	// IsStorePreferred checks if a store is in user's preferences
	IsStorePreferred(ctx context.Context, userID uuid.UUID, storeID int) (bool, error)
}

// userStorePreferenceRepository defines the repository interface (to avoid import cycle)
type userStorePreferenceRepository interface {
	GetPreferredStoreIDs(ctx context.Context, userID uuid.UUID) ([]int, error)
	GetPreferredStores(ctx context.Context, userID uuid.UUID) ([]*models.Store, error)
	SetPreferredStores(ctx context.Context, userID uuid.UUID, storeIDs []int) error
	AddPreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error
	RemovePreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error
	IsStorePreferred(ctx context.Context, userID uuid.UUID, storeID int) (bool, error)
}

type userStorePreferenceService struct {
	repo         userStorePreferenceRepository
	storeService StoreService
}

// NewUserStorePreferenceService creates a new user store preference service
func NewUserStorePreferenceService(db *bun.DB, storeService StoreService) UserStorePreferenceService {
	return &userStorePreferenceService{
		repo:         newUserStorePreferenceRepo(db),
		storeService: storeService,
	}
}

// newUserStorePreferenceRepo creates an internal repository (avoids import cycle)
func newUserStorePreferenceRepo(db *bun.DB) userStorePreferenceRepository {
	return &userStorePreferenceRepo{db: db}
}

// userStorePreferenceRepo is the internal repository implementation
type userStorePreferenceRepo struct {
	db *bun.DB
}

func (r *userStorePreferenceRepo) GetPreferredStoreIDs(ctx context.Context, userID uuid.UUID) ([]int, error) {
	var storeIDs []int
	err := r.db.NewSelect().
		Model((*models.UserStorePreference)(nil)).
		Column("store_id").
		Where("user_id = ?", userID).
		Order("created_at ASC").
		Scan(ctx, &storeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferred store IDs: %w", err)
	}
	return storeIDs, nil
}

func (r *userStorePreferenceRepo) GetPreferredStores(ctx context.Context, userID uuid.UUID) ([]*models.Store, error) {
	var preferences []*models.UserStorePreference
	err := r.db.NewSelect().
		Model(&preferences).
		Relation("Store").
		Where("usp.user_id = ?", userID).
		Order("usp.created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get preferred stores: %w", err)
	}
	stores := make([]*models.Store, 0, len(preferences))
	for _, pref := range preferences {
		if pref.Store != nil {
			stores = append(stores, pref.Store)
		}
	}
	return stores, nil
}

func (r *userStorePreferenceRepo) SetPreferredStores(ctx context.Context, userID uuid.UUID, storeIDs []int) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewDelete().
			Model((*models.UserStorePreference)(nil)).
			Where("user_id = ?", userID).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to delete existing preferences: %w", err)
		}
		if len(storeIDs) == 0 {
			return nil
		}
		now := time.Now()
		preferences := make([]*models.UserStorePreference, len(storeIDs))
		for i, storeID := range storeIDs {
			preferences[i] = &models.UserStorePreference{
				UserID:    userID,
				StoreID:   storeID,
				CreatedAt: now,
			}
		}
		_, err = tx.NewInsert().Model(&preferences).Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to insert new preferences: %w", err)
		}
		return nil
	})
}

func (r *userStorePreferenceRepo) AddPreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error {
	pref := &models.UserStorePreference{
		UserID:    userID,
		StoreID:   storeID,
		CreatedAt: time.Now(),
	}
	_, err := r.db.NewInsert().
		Model(pref).
		On("CONFLICT (user_id, store_id) DO NOTHING").
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add preferred store: %w", err)
	}
	return nil
}

func (r *userStorePreferenceRepo) RemovePreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error {
	_, err := r.db.NewDelete().
		Model((*models.UserStorePreference)(nil)).
		Where("user_id = ? AND store_id = ?", userID, storeID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to remove preferred store: %w", err)
	}
	return nil
}

func (r *userStorePreferenceRepo) IsStorePreferred(ctx context.Context, userID uuid.UUID, storeID int) (bool, error) {
	count, err := r.db.NewSelect().
		Model((*models.UserStorePreference)(nil)).
		Where("user_id = ? AND store_id = ?", userID, storeID).
		Count(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if store is preferred: %w", err)
	}
	return count > 0, nil
}

func (s *userStorePreferenceService) GetPreferredStoreIDs(ctx context.Context, userID uuid.UUID) ([]int, error) {
	storeIDs, err := s.repo.GetPreferredStoreIDs(ctx, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get preferred store IDs")
	}
	return storeIDs, nil
}

func (s *userStorePreferenceService) GetPreferredStores(ctx context.Context, userID uuid.UUID) ([]*models.Store, error) {
	stores, err := s.repo.GetPreferredStores(ctx, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get preferred stores")
	}
	return stores, nil
}

func (s *userStorePreferenceService) SetPreferredStores(ctx context.Context, userID uuid.UUID, storeIDs []int) error {
	// Validate that all store IDs exist
	if len(storeIDs) > 0 {
		stores, err := s.storeService.GetByIDs(ctx, storeIDs)
		if err != nil {
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to validate store IDs")
		}

		// Create a set of valid store IDs
		validStoreIDs := make(map[int]bool)
		for _, store := range stores {
			validStoreIDs[store.ID] = true
		}

		// Check if all requested store IDs are valid
		for _, storeID := range storeIDs {
			if !validStoreIDs[storeID] {
				return apperrors.NotFound(fmt.Sprintf("store with ID %d not found", storeID))
			}
		}
	}

	if err := s.repo.SetPreferredStores(ctx, userID, storeIDs); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to set preferred stores")
	}

	return nil
}

func (s *userStorePreferenceService) AddPreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error {
	// Validate that the store exists
	_, err := s.storeService.GetByID(ctx, storeID)
	if err != nil {
		return apperrors.NotFound(fmt.Sprintf("store with ID %d not found", storeID))
	}

	if err := s.repo.AddPreferredStore(ctx, userID, storeID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to add preferred store")
	}

	return nil
}

func (s *userStorePreferenceService) RemovePreferredStore(ctx context.Context, userID uuid.UUID, storeID int) error {
	if err := s.repo.RemovePreferredStore(ctx, userID, storeID); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to remove preferred store")
	}

	return nil
}

func (s *userStorePreferenceService) IsStorePreferred(ctx context.Context, userID uuid.UUID, storeID int) (bool, error) {
	isPreferred, err := s.repo.IsStorePreferred(ctx, userID, storeID)
	if err != nil {
		return false, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to check if store is preferred")
	}
	return isPreferred, nil
}
