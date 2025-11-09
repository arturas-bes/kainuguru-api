package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services/matching"
	"github.com/kainuguru/kainuguru-api/pkg/normalize"
	"github.com/uptrace/bun"
)

type productMasterService struct {
	db         *bun.DB
	matcher    *matching.CompositeMatcher
	normalizer *normalize.LithuanianNormalizer
	logger     *slog.Logger
}

// NewProductMasterService creates a new product master service instance
func NewProductMasterService(db *bun.DB) ProductMasterService {
	return &productMasterService{
		db:         db,
		matcher:    matching.NewCompositeMatcher(db),
		normalizer: normalize.NewLithuanianNormalizer(),
		logger:     slog.Default().With("service", "product_master"),
	}
}

// Basic CRUD operations
func (s *productMasterService) GetByID(ctx context.Context, id int64) (*models.ProductMaster, error) {
	master := &models.ProductMaster{}
	err := s.db.NewSelect().
		Model(master).
		Where("pm.id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get product master by ID %d: %w", id, err)
	}

	return master, nil
}

func (s *productMasterService) GetByIDs(ctx context.Context, ids []int64) ([]*models.ProductMaster, error) {
	if len(ids) == 0 {
		return []*models.ProductMaster{}, nil
	}

	var masters []*models.ProductMaster
	err := s.db.NewSelect().
		Model(&masters).
		Where("pm.id IN (?)", bun.In(ids)).
		Order("pm.id ASC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get product masters by IDs: %w", err)
	}

	return masters, nil
}

func (s *productMasterService) GetAll(ctx context.Context, filters ProductMasterFilters) ([]*models.ProductMaster, error) {
	query := s.db.NewSelect().Model((*models.ProductMaster)(nil))

	// Apply filters
	if len(filters.Status) > 0 {
		query = query.Where("pm.status IN (?)", bun.In(filters.Status))
	}

	if len(filters.Categories) > 0 {
		query = query.Where("pm.category IN (?)", bun.In(filters.Categories))
	}

	if len(filters.Brands) > 0 {
		query = query.Where("pm.brand IN (?)", bun.In(filters.Brands))
	}

	// Note: isVerified and isActive filters removed - these columns don't exist in DB
	// The schema uses "status" field instead

	if filters.MinConfidence != nil {
		query = query.Where("pm.confidence_score >= ?", *filters.MinConfidence)
	}

	if filters.MinMatches != nil {
		query = query.Where("pm.match_count >= ?", *filters.MinMatches)
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	// Default ordering
	query = query.Order("pm.id DESC")

	var masters []*models.ProductMaster
	err := query.Scan(ctx, &masters)
	if err != nil {
		return nil, fmt.Errorf("failed to get product masters: %w", err)
	}

	return masters, nil
}

func (s *productMasterService) Create(ctx context.Context, master *models.ProductMaster) error {
	now := time.Now()
	master.CreatedAt = now
	master.UpdatedAt = now

	if master.Status == "" {
		master.Status = string(models.ProductMasterStatusActive)
	}

	if master.NormalizedName == "" {
		master.NormalizeName()
	}

	_, err := s.db.NewInsert().
		Model(master).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to create product master: %w", err)
	}

	s.logger.Info("product master created",
		slog.Int64("id", master.ID),
		slog.String("name", master.Name),
	)

	return nil
}

func (s *productMasterService) Update(ctx context.Context, master *models.ProductMaster) error {
	master.UpdatedAt = time.Now()

	if master.NormalizedName == "" {
		master.NormalizeName()
	}

	result, err := s.db.NewUpdate().
		Model(master).
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to update product master: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product master not found: %d", master.ID)
	}

	s.logger.Info("product master updated",
		slog.Int64("id", master.ID),
		slog.String("name", master.Name),
	)

	return nil
}

func (s *productMasterService) Delete(ctx context.Context, id int64) error {
	master := &models.ProductMaster{
		ID:        id,
		Status:    string(models.ProductMasterStatusDeleted),
		UpdatedAt: time.Now(),
	}

	result, err := s.db.NewUpdate().
		Model(master).
		Column("status", "updated_at").
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to delete product master: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product master not found: %d", id)
	}

	s.logger.Info("product master deleted",
		slog.Int64("id", id),
	)

	return nil
}

// Product master operations
func (s *productMasterService) GetByCanonicalName(ctx context.Context, name string) (*models.ProductMaster, error) {
	normalizedName := s.normalizer.NormalizeForSearch(name)

	master := &models.ProductMaster{}
	err := s.db.NewSelect().
		Model(master).
		Where("pm.normalized_name = ?", normalizedName).
		Where("pm.status = ?", models.ProductMasterStatusActive).
		Order("pm.confidence_score DESC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get product master by canonical name: %w", err)
	}

	return master, nil
}

func (s *productMasterService) GetActiveProductMasters(ctx context.Context) ([]*models.ProductMaster, error) {
	var masters []*models.ProductMaster
	err := s.db.NewSelect().
		Model(&masters).
		Where("pm.status = ?", models.ProductMasterStatusActive).
		Order("pm.match_count DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get active product masters: %w", err)
	}

	return masters, nil
}

func (s *productMasterService) GetVerifiedProductMasters(ctx context.Context) ([]*models.ProductMaster, error) {
	var masters []*models.ProductMaster
	err := s.db.NewSelect().
		Model(&masters).
		Where("pm.status = ?", models.ProductMasterStatusActive).
		Where("pm.confidence_score >= ?", 0.8).
		Order("pm.match_count DESC").
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get verified product masters: %w", err)
	}

	return masters, nil
}

func (s *productMasterService) GetProductMastersForReview(ctx context.Context) ([]*models.ProductMaster, error) {
	var masters []*models.ProductMaster
	err := s.db.NewSelect().
		Model(&masters).
		Where("pm.status = ?", models.ProductMasterStatusActive).
		Where("pm.confidence_score < ?", 0.8).
		Where("pm.confidence_score >= ?", 0.3).
		Order("pm.created_at DESC").
		Limit(100).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get product masters for review: %w", err)
	}

	return masters, nil
}

// Matching operations
func (s *productMasterService) FindMatchingMasters(ctx context.Context, productName string, brand string, category string) ([]*models.ProductMaster, error) {
	matches, err := s.FindMatchingMastersWithScores(ctx, productName, brand, category)
	if err != nil {
		return nil, err
	}

	var masters []*models.ProductMaster
	for _, match := range matches {
		masters = append(masters, match.Master)
	}

	return masters, nil
}

func (s *productMasterService) FindMatchingMastersWithScores(ctx context.Context, productName string, brand string, category string) ([]*ProductMasterMatch, error) {
	if strings.TrimSpace(productName) == "" {
		return nil, fmt.Errorf("product name cannot be empty")
	}

	product := &models.Product{
		Name:           productName,
		NormalizedName: s.normalizer.NormalizeForSearch(productName),
	}

	if brand != "" {
		product.Brand = &brand
	}

	if category != "" {
		product.Category = &category
	}

	matchResults, err := s.matcher.FindBestMatches(ctx, product, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to find matching masters: %w", err)
	}

	var matches []*ProductMasterMatch
	for _, result := range matchResults {
		s.logger.Debug("found match",
			slog.String("product", productName),
			slog.String("master", result.Master.Name),
			slog.Float64("score", result.Score),
			slog.String("method", result.Method),
		)
		matches = append(matches, &ProductMasterMatch{
			Master:     result.Master,
			MatchScore: result.Score,
			Method:     result.Method,
		})
	}

	return matches, nil
}

func (s *productMasterService) MatchProduct(ctx context.Context, productID int, masterID int64) error {
	product := &models.Product{}
	err := s.db.NewSelect().
		Model(product).
		Where("p.id = ?", productID).
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	master := &models.ProductMaster{}
	err = s.db.NewSelect().
		Model(master).
		Where("pm.id = ?", masterID).
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("failed to get product master: %w", err)
	}

	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewUpdate().
			Model((*models.Product)(nil)).
			Set("product_master_id = ?", masterID).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", productID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}

		master.IncrementMatchCount()
		_, err = tx.NewUpdate().
			Model(master).
			Column("match_count", "last_seen_date", "updated_at").
			WherePK().
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to update master: %w", err)
		}

		_, err = tx.NewInsert().
			Model(&models.ProductMasterMatch{
				ProductID:       int64(productID),
				ProductMasterID: masterID,
				Confidence:      master.ConfidenceScore,
				MatchType:       "manual",
				ReviewStatus:    "approved",
			}).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to create match record: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Info("product matched to master",
		slog.Int("product_id", productID),
		slog.Int64("master_id", masterID),
	)

	return nil
}

// FindBestMatch finds the best matching product masters for a product
func (s *productMasterService) FindBestMatch(ctx context.Context, product *models.Product, limit int) ([]*ProductMasterMatch, error) {
	matchResults, err := s.matcher.FindBestMatches(ctx, product, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find best matches: %w", err)
	}

	var matches []*ProductMasterMatch
	for _, result := range matchResults {
		matches = append(matches, &ProductMasterMatch{
			Master:      result.Master,
			MatchScore:  result.Score,
			Method:      result.Method,
			Confidence:  result.Confidence,
		})
	}

	return matches, nil
}

// CreateFromProduct creates a new product master from a product
func (s *productMasterService) CreateFromProduct(ctx context.Context, product *models.Product) (*models.ProductMaster, error) {
	now := time.Now()
	
	// Normalize product name by removing brand
	genericName := s.normalizeProductName(product.Name, product.Brand)
	
	master := &models.ProductMaster{
		Name:            genericName,
		NormalizedName:  s.normalizer.NormalizeForSearch(genericName),
		Brand:           product.Brand,
		Category:        product.Category,
		Subcategory:     product.Subcategory,
		Tags:            product.Tags,
		MatchCount:      1,
		ConfidenceScore: 0.5,
		LastSeenDate:    &now,
		Status:          string(models.ProductMasterStatusActive),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if product.UnitSize != nil {
		master.StandardUnit = product.UnitSize
	}

	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().
			Model(master).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to create master: %w", err)
		}

		_, err = tx.NewInsert().
			Model(&models.ProductMasterMatch{
				ProductID:       int64(product.ID),
				ProductMasterID: master.ID,
				Confidence:      master.ConfidenceScore,
				MatchType:       "new_master",
				ReviewStatus:    "pending",
			}).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to create match record: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("created master from product",
		slog.Int("product_id", product.ID),
		slog.Int64("master_id", master.ID),
		slog.String("name", master.Name),
		slog.String("original_name", product.Name),
	)

	return master, nil
}

func (s *productMasterService) CreateMasterFromProduct(ctx context.Context, productID int) (*models.ProductMaster, error) {
	product := &models.Product{}
	err := s.db.NewSelect().
		Model(product).
		Relation("Store").
		Where("p.id = ?", productID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	now := time.Now()
	
	// Normalize product name by removing brand
	genericName := s.normalizeProductName(product.Name, product.Brand)
	
	master := &models.ProductMaster{
		Name:            genericName,
		NormalizedName:  s.normalizer.NormalizeForSearch(genericName),
		Brand:           product.Brand,
		Category:        product.Category,
		Subcategory:     product.Subcategory,
		Tags:            product.Tags,
		StandardUnit:    product.UnitType,
		MatchCount:      1,
		ConfidenceScore: 0.5,
		LastSeenDate:    &now,
		Status:          string(models.ProductMasterStatusActive),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if product.UnitSize != nil {
		master.StandardUnit = product.UnitSize
	}

	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewInsert().
			Model(master).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to create master: %w", err)
		}

		_, err = tx.NewUpdate().
			Model((*models.Product)(nil)).
			Set("product_master_id = ?", master.ID).
			Set("updated_at = ?", now).
			Where("id = ?", productID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to link product: %w", err)
		}

		_, err = tx.NewInsert().
			Model(&models.ProductMasterMatch{
				ProductID:       int64(productID),
				ProductMasterID: master.ID,
				Confidence:      master.ConfidenceScore,
				MatchType:       "new_master",
				ReviewStatus:    "pending",
			}).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to create match record: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	s.logger.Info("created master from product",
		slog.Int("product_id", productID),
		slog.Int64("master_id", master.ID),
		slog.String("name", master.Name),
	)

	return master, nil
}

// Verification operations
func (s *productMasterService) VerifyProductMaster(ctx context.Context, masterID int64, verifierID string) error {
	master := &models.ProductMaster{}
	err := s.db.NewSelect().
		Model(master).
		Where("pm.id = ?", masterID).
		Scan(ctx)

	if err != nil {
		return fmt.Errorf("failed to get product master: %w", err)
	}

	master.ConfidenceScore = 1.0
	master.UpdatedAt = time.Now()

	_, err = s.db.NewUpdate().
		Model(master).
		Column("confidence_score", "updated_at").
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to verify product master: %w", err)
	}

	s.logger.Info("product master verified",
		slog.Int64("master_id", masterID),
		slog.String("verifier", verifierID),
	)

	return nil
}

func (s *productMasterService) DeactivateProductMaster(ctx context.Context, masterID int64) error {
	master := &models.ProductMaster{
		ID:        masterID,
		Status:    string(models.ProductMasterStatusInactive),
		UpdatedAt: time.Now(),
	}

	result, err := s.db.NewUpdate().
		Model(master).
		Column("status", "updated_at").
		WherePK().
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to deactivate product master: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product master not found: %d", masterID)
	}

	s.logger.Info("product master deactivated",
		slog.Int64("master_id", masterID),
	)

	return nil
}

func (s *productMasterService) MarkAsDuplicate(ctx context.Context, masterID int64, duplicateOfID int64) error {
	err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		master := &models.ProductMaster{}
		err := tx.NewSelect().
			Model(master).
			Where("pm.id = ?", masterID).
			For("UPDATE").
			Scan(ctx)

		if err != nil {
			return fmt.Errorf("failed to get product master: %w", err)
		}

		master.MarkAsMerged(duplicateOfID)

		_, err = tx.NewUpdate().
			Model(master).
			Column("status", "merged_into_id", "updated_at").
			WherePK().
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to mark as duplicate: %w", err)
		}

		_, err = tx.NewUpdate().
			Model((*models.Product)(nil)).
			Set("product_master_id = ?", duplicateOfID).
			Set("updated_at = ?", time.Now()).
			Where("product_master_id = ?", masterID).
			Exec(ctx)

		if err != nil {
			return fmt.Errorf("failed to reassign products: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Info("product master marked as duplicate",
		slog.Int64("master_id", masterID),
		slog.Int64("duplicate_of_id", duplicateOfID),
	)

	return nil
}

// Statistics
func (s *productMasterService) GetMatchingStatistics(ctx context.Context, masterID int64) (*ProductMasterStats, error) {
	master := &models.ProductMaster{}
	err := s.db.NewSelect().
		Model(master).
		Where("pm.id = ?", masterID).
		Scan(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to get product master: %w", err)
	}

	productCount, err := s.db.NewSelect().
		Model((*models.Product)(nil)).
		Where("product_master_id = ?", masterID).
		Count(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	stats := &ProductMasterStats{
		TotalMatches:      master.MatchCount,
		SuccessfulMatches: productCount,
		FailedMatches:     0,
		SuccessRate:       100.0,
		ConfidenceScore:   master.ConfidenceScore,
		LastMatchedAt:     master.LastSeenDate,
	}

	return stats, nil
}

func (s *productMasterService) GetOverallMatchingStats(ctx context.Context) (*OverallMatchingStats, error) {
	var stats OverallMatchingStats

	totalMasters, err := s.db.NewSelect().
		Model((*models.ProductMaster)(nil)).
		Where("status = ?", models.ProductMasterStatusActive).
		Count(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to count masters: %w", err)
	}

	verifiedMasters, err := s.db.NewSelect().
		Model((*models.ProductMaster)(nil)).
		Where("status = ?", models.ProductMasterStatusActive).
		Where("confidence_score >= ?", 0.8).
		Count(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to count verified masters: %w", err)
	}

	totalProducts, err := s.db.NewSelect().
		Model((*models.Product)(nil)).
		Count(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	matchedProducts, err := s.db.NewSelect().
		Model((*models.Product)(nil)).
		Where("product_master_id IS NOT NULL").
		Count(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to count matched products: %w", err)
	}

	stats.TotalProducts = totalProducts
	stats.MatchedProducts = matchedProducts
	stats.UnmatchedProducts = totalProducts - matchedProducts
	stats.ProductMasters = totalMasters
	stats.VerifiedMasters = verifiedMasters

	if totalProducts > 0 {
		stats.OverallMatchRate = float64(matchedProducts) / float64(totalProducts)
	}

	if totalMasters > 0 {
		err = s.db.NewSelect().
			Model((*models.ProductMaster)(nil)).
			ColumnExpr("AVG(confidence_score)").
			Where("status = ?", models.ProductMasterStatusActive).
			Scan(ctx, &stats.AverageConfidence)
		if err != nil {
			s.logger.Error("failed to calculate average confidence", slog.String("error", err.Error()))
		}
	}

	return &stats, nil
}

// normalizeProductName removes brand names from product names to create generic product masters
// Examples:
// "Saulėgrąžų aliejus NATURA 1L" -> "Saulėgrąžų aliejus"
// "Glaistytas varškės sūrelis MAGIJA" -> "Glaistytas varškės sūrelis"
// "SOSTINĖS batonas" -> "Batonas"
// "IKI varškė" -> "Varškė"
func (s *productMasterService) normalizeProductName(name string, brand *string) string {
	normalized := name
	
	// Remove brand from name if present
	if brand != nil && *brand != "" {
		brandUpper := strings.ToUpper(*brand)
		brandLower := strings.ToLower(*brand)
		brandTitle := strings.Title(strings.ToLower(*brand))
		
		// Remove all variations of the brand
		normalized = strings.ReplaceAll(normalized, brandUpper, "")
		normalized = strings.ReplaceAll(normalized, *brand, "")
		normalized = strings.ReplaceAll(normalized, brandLower, "")
		normalized = strings.ReplaceAll(normalized, brandTitle, "")
	}
	
	// Define known brands to remove
	knownBrands := map[string]bool{
		"IKI": true, "MAXIMA": true, "RIMI": true,
		"DVARO": true, "ROKIŠKIO": true, "SVALYA": true,
		"NATURA": true, "MAGIJA": true, "SOSTINĖS": true,
		"CLEVER": true, "TARCZYNSKI": true, "JUBILIEJAUS": true,
	}
	
	words := strings.Fields(normalized)
	filteredWords := []string{}
	
	for _, word := range words {
		// Keep word if it's not all uppercase or if it's a measurement
		isAllUpper := word == strings.ToUpper(word) && len(word) > 1
		isMeasurement := strings.Contains(strings.ToLower(word), "kg") || 
						 strings.Contains(strings.ToLower(word), "ml") ||
						 strings.Contains(strings.ToLower(word), "vnt") ||
						 strings.Contains(strings.ToLower(word), "l") ||
						 strings.Contains(strings.ToLower(word), "g")
		
		if !isAllUpper || isMeasurement {
			// Not all uppercase, keep it
			filteredWords = append(filteredWords, word)
		} else {
			// All uppercase - check if it's a known brand
			if !knownBrands[strings.ToUpper(word)] {
				// Not a known brand, keep it
				filteredWords = append(filteredWords, word)
			}
			// If it's a known brand, skip it (don't add to filteredWords)
		}
	}
	
	normalized = strings.Join(filteredWords, " ")
	
	// Clean up extra spaces and punctuation
	normalized = strings.TrimSpace(normalized)
	normalized = strings.Trim(normalized, ",")
	normalized = strings.TrimSpace(normalized)
	
	// Capitalize first letter
	if len(normalized) > 0 {
		runes := []rune(normalized)
		runes[0] = unicode.ToUpper(runes[0])
		normalized = string(runes)
	}
	
	return normalized
}
