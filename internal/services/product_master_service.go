package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/productmaster"
	"github.com/kainuguru/kainuguru-api/internal/services/matching"
	"github.com/kainuguru/kainuguru-api/pkg/normalize"
	"github.com/uptrace/bun"
)

type productMasterService struct {
	db         *bun.DB
	repo       productmaster.Repository
	matcher    *matching.CompositeMatcher
	normalizer *normalize.LithuanianNormalizer
	logger     *slog.Logger
}

// NewProductMasterService creates a new product master service instance
func NewProductMasterService(db *bun.DB) ProductMasterService {
	return NewProductMasterServiceWithRepository(db, newProductMasterRepository(db))
}

// NewProductMasterServiceWithRepository allows injecting a custom repository implementation.
func NewProductMasterServiceWithRepository(db *bun.DB, repo productmaster.Repository) ProductMasterService {
	if repo == nil {
		panic("product master repository cannot be nil")
	}
	return &productMasterService{
		db:         db,
		repo:       repo,
		matcher:    matching.NewCompositeMatcher(db),
		normalizer: normalize.NewLithuanianNormalizer(),
		logger:     slog.Default().With("service", "product_master"),
	}
}

// Basic CRUD operations
func (s *productMasterService) GetByID(ctx context.Context, id int64) (*models.ProductMaster, error) {
	master, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product master by ID %d: %w", id, err)
	}

	return master, nil
}

func (s *productMasterService) GetByIDs(ctx context.Context, ids []int64) ([]*models.ProductMaster, error) {
	masters, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get product masters by IDs: %w", err)
	}

	return masters, nil
}

func (s *productMasterService) GetAll(ctx context.Context, filters ProductMasterFilters) ([]*models.ProductMaster, error) {
	f := filters
	masters, err := s.repo.GetAll(ctx, &f)
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

	if err := s.repo.Create(ctx, master); err != nil {
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

	rowsAffected, err := s.repo.Update(ctx, master)
	if err != nil {
		return fmt.Errorf("failed to update product master: %w", err)
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
	rowsAffected, err := s.repo.SoftDelete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete product master: %w", err)
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

	master, err := s.repo.GetByCanonicalName(ctx, normalizedName)
	if err != nil {
		return nil, fmt.Errorf("failed to get product master by canonical name: %w", err)
	}

	return master, nil
}

func (s *productMasterService) GetActiveProductMasters(ctx context.Context) ([]*models.ProductMaster, error) {
	masters, err := s.repo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active product masters: %w", err)
	}

	return masters, nil
}

func (s *productMasterService) GetVerifiedProductMasters(ctx context.Context) ([]*models.ProductMaster, error) {
	masters, err := s.repo.GetVerified(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get verified product masters: %w", err)
	}

	return masters, nil
}

func (s *productMasterService) GetProductMastersForReview(ctx context.Context) ([]*models.ProductMaster, error) {
	masters, err := s.repo.GetForReview(ctx)
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
	if err := s.repo.MatchProduct(ctx, productID, masterID); err != nil {
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
			Master:     result.Master,
			MatchScore: result.Score,
			Method:     result.Method,
			Confidence: result.Confidence,
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

	if err := s.repo.CreateMasterWithMatch(ctx, product, master); err != nil {
		return nil, fmt.Errorf("failed to create master from product: %w", err)
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
	product, err := s.repo.GetProduct(ctx, productID)
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

	if err := s.repo.CreateMasterAndLinkProduct(ctx, product, master); err != nil {
		return nil, fmt.Errorf("failed to create master: %w", err)
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
	if err := s.repo.VerifyMaster(ctx, masterID, 1.0, time.Now()); err != nil {
		return fmt.Errorf("failed to verify product master: %w", err)
	}

	s.logger.Info("product master verified",
		slog.Int64("master_id", masterID),
		slog.String("verifier", verifierID),
	)

	return nil
}

func (s *productMasterService) DeactivateProductMaster(ctx context.Context, masterID int64) error {
	rowsAffected, err := s.repo.DeactivateMaster(ctx, masterID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to deactivate product master: %w", err)
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
	if err := s.repo.MarkAsDuplicate(ctx, masterID, duplicateOfID); err != nil {
		return fmt.Errorf("failed to mark as duplicate: %w", err)
	}

	s.logger.Info("product master marked as duplicate",
		slog.Int64("master_id", masterID),
		slog.Int64("duplicate_of_id", duplicateOfID),
	)

	return nil
}

// Statistics
func (s *productMasterService) GetMatchingStatistics(ctx context.Context, masterID int64) (*ProductMasterStats, error) {
	stats, err := s.repo.GetMatchingStatistics(ctx, masterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get matching statistics: %w", err)
	}

	successRate := 100.0
	if stats.TotalMatches > 0 && stats.ProductCount < stats.TotalMatches {
		successRate = (float64(stats.ProductCount) / float64(stats.TotalMatches)) * 100
	}

	return &ProductMasterStats{
		TotalMatches:      stats.TotalMatches,
		SuccessfulMatches: stats.ProductCount,
		FailedMatches:     stats.TotalMatches - stats.ProductCount,
		SuccessRate:       successRate,
		ConfidenceScore:   stats.Confidence,
		LastMatchedAt:     stats.LastMatchedAt,
	}, nil
}

func (s *productMasterService) GetOverallMatchingStats(ctx context.Context) (*OverallMatchingStats, error) {
	stats, err := s.repo.GetOverallStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	result := &OverallMatchingStats{
		TotalProducts:     stats.TotalProducts,
		MatchedProducts:   stats.MatchedProducts,
		UnmatchedProducts: stats.TotalProducts - stats.MatchedProducts,
		ProductMasters:    stats.TotalMasters,
		VerifiedMasters:   stats.VerifiedMasters,
		AverageConfidence: stats.AverageConfidence,
	}

	if stats.TotalProducts > 0 {
		result.OverallMatchRate = float64(stats.MatchedProducts) / float64(stats.TotalProducts)
	}

	return result, nil
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
