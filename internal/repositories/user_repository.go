package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/repositories/base"
	"github.com/uptrace/bun"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// User management operations
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error)
	GetAll(ctx context.Context, filters *UserFilters) ([]*models.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateEmailVerification(ctx context.Context, userID uuid.UUID, verified bool) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error
	UpdateMetadata(ctx context.Context, userID uuid.UUID, metadata models.UserMetadata) error

	// Status operations
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	GetActiveUsers(ctx context.Context, limit int) ([]*models.User, error)

	// Search and filtering
	SearchUsers(ctx context.Context, query string, filters *UserFilters) ([]*models.User, error)
	GetUsersByCreatedDate(ctx context.Context, from, to time.Time) ([]*models.User, error)
	GetUnverifiedUsers(ctx context.Context, olderThan time.Time) ([]*models.User, error)

	// Statistics
	GetUserCount(ctx context.Context) (int, error)
	GetActiveUserCount(ctx context.Context) (int, error)
	GetVerifiedUserCount(ctx context.Context) (int, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db   *bun.DB
	base *base.Repository[models.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{
		db:   db,
		base: base.NewRepository[models.User](db, "u.id"),
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := r.base.Create(ctx, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := r.base.GetByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.NewSelect().
		Model(user).
		Where("u.email = ?", email).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	if err := r.base.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.base.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetByIDs retrieves multiple users by their IDs
func (r *userRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error) {
	if len(ids) == 0 {
		return []*models.User{}, nil
	}

	users, err := r.base.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}

	return users, nil
}

// GetAll retrieves users with optional filtering
func (r *userRepository) GetAll(ctx context.Context, filters *UserFilters) ([]*models.User, error) {
	users, err := r.base.GetAll(ctx, base.WithQuery[models.User](func(q *bun.SelectQuery) *bun.SelectQuery {
		q = applyUserFilters(q, filters)
		return applyUserOrdering(q, filters)
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	return users, nil
}

// UpdatePassword updates a user's password hash
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_hash = ?", passwordHash).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateEmailVerification updates a user's email verification status
func (r *userRepository) UpdateEmailVerification(ctx context.Context, userID uuid.UUID, verified bool) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("email_verified = ?", verified).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update email verification: %w", err)
	}

	return nil
}

// UpdateLastLogin updates a user's last login timestamp
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("last_login_at = ?", loginTime).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// UpdateMetadata updates a user's metadata
func (r *userRepository) UpdateMetadata(ctx context.Context, userID uuid.UUID, metadata models.UserMetadata) error {
	// Create a temporary user to marshal metadata
	tempUser := &models.User{ID: userID}
	if err := tempUser.SetMetadata(metadata); err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("metadata = ?", tempUser.MetadataJSON).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

// ActivateUser activates a user account
func (r *userRepository) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_active = ?", true).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to activate user: %w", err)
	}

	return nil
}

// DeactivateUser deactivates a user account
func (r *userRepository) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("is_active = ?", false).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	return nil
}

// GetActiveUsers retrieves active users with a limit
func (r *userRepository) GetActiveUsers(ctx context.Context, limit int) ([]*models.User, error) {
	query := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("u.is_active = ?", true).
		Order("u.last_login_at DESC NULLS LAST")

	if limit > 0 {
		query = query.Limit(limit)
	}

	var users []*models.User
	err := query.Scan(ctx, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	return users, nil
}

// SearchUsers searches users by email or full name
func (r *userRepository) SearchUsers(ctx context.Context, query string, filters *UserFilters) ([]*models.User, error) {
	dbQuery := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("(u.email ILIKE ? OR u.full_name ILIKE ?)", "%"+query+"%", "%"+query+"%")

	// Apply additional filters
	if filters != nil {
		if filters.IsActive != nil {
			dbQuery = dbQuery.Where("u.is_active = ?", *filters.IsActive)
		}

		if filters.IsVerified != nil {
			dbQuery = dbQuery.Where("u.email_verified = ?", *filters.IsVerified)
		}

		if filters.Limit > 0 {
			dbQuery = dbQuery.Limit(filters.Limit)
		}

		if filters.Offset > 0 {
			dbQuery = dbQuery.Offset(filters.Offset)
		}
	}

	// Default ordering by relevance (email match first, then name)
	dbQuery = dbQuery.Order("CASE WHEN u.email ILIKE ? THEN 1 ELSE 2 END, u.created_at DESC", "%"+query+"%")

	var users []*models.User
	err := dbQuery.Scan(ctx, &users)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}

// GetUsersByCreatedDate retrieves users created within a date range
func (r *userRepository) GetUsersByCreatedDate(ctx context.Context, from, to time.Time) ([]*models.User, error) {
	var users []*models.User
	err := r.db.NewSelect().
		Model(&users).
		Where("u.created_at >= ? AND u.created_at <= ?", from, to).
		Order("u.created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get users by created date: %w", err)
	}

	return users, nil
}

// GetUnverifiedUsers retrieves users who haven't verified their email
func (r *userRepository) GetUnverifiedUsers(ctx context.Context, olderThan time.Time) ([]*models.User, error) {
	var users []*models.User
	err := r.db.NewSelect().
		Model(&users).
		Where("u.email_verified = ? AND u.created_at < ?", false, olderThan).
		Order("u.created_at ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get unverified users: %w", err)
	}

	return users, nil
}

// GetUserCount returns the total number of users
func (r *userRepository) GetUserCount(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get user count: %w", err)
	}

	return count, nil
}

// GetActiveUserCount returns the number of active users
func (r *userRepository) GetActiveUserCount(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("u.is_active = ?", true).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get active user count: %w", err)
	}

	return count, nil
}

// GetVerifiedUserCount returns the number of verified users
func (r *userRepository) GetVerifiedUserCount(ctx context.Context) (int, error) {
	count, err := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("u.email_verified = ?", true).
		Count(ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to get verified user count: %w", err)
	}

	return count, nil
}

// UserFilters represents filters for user queries
type UserFilters struct {
	IsActive          *bool      `json:"isActive"`
	IsVerified        *bool      `json:"isVerified"`
	PreferredLanguage string     `json:"preferredLanguage"`
	HasOAuth          *bool      `json:"hasOAuth"`
	CreatedAfter      *time.Time `json:"createdAfter"`
	CreatedBefore     *time.Time `json:"createdBefore"`
	LastLoginAfter    *time.Time `json:"lastLoginAfter"`
	Limit             int        `json:"limit"`
	Offset            int        `json:"offset"`
	OrderBy           string     `json:"orderBy"`
	OrderDir          string     `json:"orderDir"`
}

func applyUserFilters(query *bun.SelectQuery, filters *UserFilters) *bun.SelectQuery {
	if filters == nil {
		return query
	}
	if filters.IsActive != nil {
		query = query.Where("u.is_active = ?", *filters.IsActive)
	}
	if filters.IsVerified != nil {
		query = query.Where("u.email_verified = ?", *filters.IsVerified)
	}
	if filters.PreferredLanguage != "" {
		query = query.Where("u.preferred_language = ?", filters.PreferredLanguage)
	}
	if filters.CreatedAfter != nil {
		query = query.Where("u.created_at > ?", *filters.CreatedAfter)
	}
	if filters.CreatedBefore != nil {
		query = query.Where("u.created_at < ?", *filters.CreatedBefore)
	}
	if filters.LastLoginAfter != nil {
		query = query.Where("u.last_login_at > ?", *filters.LastLoginAfter)
	}
	if filters.HasOAuth != nil {
		if *filters.HasOAuth {
			query = query.Where("u.oauth_provider IS NOT NULL")
		} else {
			query = query.Where("u.oauth_provider IS NULL")
		}
	}
	return query
}

func applyUserOrdering(query *bun.SelectQuery, filters *UserFilters) *bun.SelectQuery {
	if filters == nil {
		return query.Order("u.created_at DESC")
	}
	orderBy := "u.created_at"
	if filters.OrderBy != "" {
		orderBy = fmt.Sprintf("u.%s", filters.OrderBy)
	}
	orderDir := "DESC"
	if filters.OrderDir == "ASC" {
		orderDir = "ASC"
	}
	query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDir))
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}
	return query
}
