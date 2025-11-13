package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/productmaster"
	"github.com/kainuguru/kainuguru-api/pkg/normalize"
)

func TestProductMasterService_GetAllDelegates(t *testing.T) {
	repo := &productMasterRepoStub{
		getAllFunc: func(ctx context.Context, filters *productmaster.Filters) ([]*models.ProductMaster, error) {
			if filters == nil || filters.Limit != 10 {
				t.Fatalf("filters not forwarded: %+v", filters)
			}
			return []*models.ProductMaster{{ID: 1}}, nil
		},
	}
	svc := &productMasterService{repo: repo}

	rs, err := svc.GetAll(context.Background(), ProductMasterFilters{Limit: 10})
	if err != nil {
		t.Fatalf("GetAll returned error: %v", err)
	}
	if len(rs) != 1 {
		t.Fatalf("unexpected masters: %+v", rs)
	}
}

func TestProductMasterService_CreateUsesRepo(t *testing.T) {
	repo := &productMasterRepoStub{createFunc: func(ctx context.Context, master *models.ProductMaster) error {
		if master.NormalizedName == "" {
			t.Fatalf("expected normalized name")
		}
		return nil
	}}
	svc := &productMasterService{repo: repo, logger: noopLogger(), normalizer: normalize.NewLithuanianNormalizer()}

	master := &models.ProductMaster{Name: "Test"}
	if err := svc.Create(context.Background(), master); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
}

func TestProductMasterService_UpdatePropagatesErrors(t *testing.T) {
	want := errors.New("boom")
	repo := &productMasterRepoStub{updateFunc: func(ctx context.Context, master *models.ProductMaster) (int64, error) {
		return 0, want
	}}
	svc := &productMasterService{repo: repo, logger: noopLogger(), normalizer: normalize.NewLithuanianNormalizer()}

	if err := svc.Update(context.Background(), &models.ProductMaster{Name: "Test"}); !errors.Is(err, want) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}

type productMasterRepoStub struct {
	getByIDFunc      func(ctx context.Context, id int64) (*models.ProductMaster, error)
	getByIDsFunc     func(ctx context.Context, ids []int64) ([]*models.ProductMaster, error)
	getAllFunc       func(ctx context.Context, filters *productmaster.Filters) ([]*models.ProductMaster, error)
	createFunc       func(ctx context.Context, master *models.ProductMaster) error
	updateFunc       func(ctx context.Context, master *models.ProductMaster) (int64, error)
	softDeleteFunc   func(ctx context.Context, id int64) (int64, error)
	getCanonicalFunc func(ctx context.Context, normalizedName string) (*models.ProductMaster, error)
	getActiveFunc    func(ctx context.Context) ([]*models.ProductMaster, error)
	getVerifiedFunc  func(ctx context.Context) ([]*models.ProductMaster, error)
	getForReviewFunc func(ctx context.Context) ([]*models.ProductMaster, error)
	matchProductFunc func(ctx context.Context, productID int, masterID int64) error
	getStatsFunc     func(ctx context.Context, masterID int64) (*productmaster.ProductMasterStats, error)
	getOverallFunc   func(ctx context.Context) (*productmaster.OverallStats, error)
	createMasterFunc func(ctx context.Context, product *models.Product, master *models.ProductMaster) error
	getProductFunc   func(ctx context.Context, productID int) (*models.Product, error)
}

func (s *productMasterRepoStub) GetByID(ctx context.Context, id int64) (*models.ProductMaster, error) {
	if s.getByIDFunc != nil {
		return s.getByIDFunc(ctx, id)
	}
	return &models.ProductMaster{}, nil
}

func (s *productMasterRepoStub) GetByIDs(ctx context.Context, ids []int64) ([]*models.ProductMaster, error) {
	if s.getByIDsFunc != nil {
		return s.getByIDsFunc(ctx, ids)
	}
	return []*models.ProductMaster{}, nil
}

func (s *productMasterRepoStub) GetAll(ctx context.Context, filters *productmaster.Filters) ([]*models.ProductMaster, error) {
	if s.getAllFunc != nil {
		return s.getAllFunc(ctx, filters)
	}
	return []*models.ProductMaster{}, nil
}

func (s *productMasterRepoStub) Create(ctx context.Context, master *models.ProductMaster) error {
	if s.createFunc != nil {
		return s.createFunc(ctx, master)
	}
	return nil
}

func (s *productMasterRepoStub) Update(ctx context.Context, master *models.ProductMaster) (int64, error) {
	if s.updateFunc != nil {
		return s.updateFunc(ctx, master)
	}
	return 1, nil
}

func (s *productMasterRepoStub) SoftDelete(ctx context.Context, id int64) (int64, error) {
	if s.softDeleteFunc != nil {
		return s.softDeleteFunc(ctx, id)
	}
	return 1, nil
}

func (s *productMasterRepoStub) GetByCanonicalName(ctx context.Context, normalizedName string) (*models.ProductMaster, error) {
	if s.getCanonicalFunc != nil {
		return s.getCanonicalFunc(ctx, normalizedName)
	}
	return &models.ProductMaster{}, nil
}

func (s *productMasterRepoStub) GetActive(ctx context.Context) ([]*models.ProductMaster, error) {
	if s.getActiveFunc != nil {
		return s.getActiveFunc(ctx)
	}
	return []*models.ProductMaster{}, nil
}

func (s *productMasterRepoStub) GetVerified(ctx context.Context) ([]*models.ProductMaster, error) {
	if s.getVerifiedFunc != nil {
		return s.getVerifiedFunc(ctx)
	}
	return []*models.ProductMaster{}, nil
}

func (s *productMasterRepoStub) GetForReview(ctx context.Context) ([]*models.ProductMaster, error) {
	if s.getForReviewFunc != nil {
		return s.getForReviewFunc(ctx)
	}
	return []*models.ProductMaster{}, nil
}

func (s *productMasterRepoStub) MatchProduct(ctx context.Context, productID int, masterID int64) error {
	if s.matchProductFunc != nil {
		return s.matchProductFunc(ctx, productID, masterID)
	}
	return nil
}

func (s *productMasterRepoStub) VerifyMaster(ctx context.Context, masterID int64, confidence float64, verifiedAt time.Time) error {
	return nil
}

func (s *productMasterRepoStub) DeactivateMaster(ctx context.Context, masterID int64, deactivatedAt time.Time) (int64, error) {
	return 1, nil
}

func (s *productMasterRepoStub) MarkAsDuplicate(ctx context.Context, masterID, duplicateOfID int64) error {
	return nil
}

func (s *productMasterRepoStub) CreateMasterWithMatch(ctx context.Context, product *models.Product, master *models.ProductMaster) error {
	if s.createMasterFunc != nil {
		return s.createMasterFunc(ctx, product, master)
	}
	return nil
}

func (s *productMasterRepoStub) CreateMasterAndLinkProduct(ctx context.Context, product *models.Product, master *models.ProductMaster) error {
	if s.createMasterFunc != nil {
		return s.createMasterFunc(ctx, product, master)
	}
	return nil
}

func (s *productMasterRepoStub) GetProduct(ctx context.Context, productID int) (*models.Product, error) {
	if s.getProductFunc != nil {
		return s.getProductFunc(ctx, productID)
	}
	return &models.Product{}, nil
}

func (s *productMasterRepoStub) GetMatchingStatistics(ctx context.Context, masterID int64) (*productmaster.ProductMasterStats, error) {
	if s.getStatsFunc != nil {
		return s.getStatsFunc(ctx, masterID)
	}
	return &productmaster.ProductMasterStats{
		Master: &models.ProductMaster{MatchCount: 1},
	}, nil
}

func (s *productMasterRepoStub) GetOverallStatistics(ctx context.Context) (*productmaster.OverallStats, error) {
	if s.getOverallFunc != nil {
		return s.getOverallFunc(ctx)
	}
	return &productmaster.OverallStats{}, nil
}

func (s *productMasterRepoStub) GetUnmatchedProducts(ctx context.Context, limit int) ([]*models.Product, error) {
	return []*models.Product{}, nil
}

func (s *productMasterRepoStub) MarkProductForReview(ctx context.Context, productID int) error {
	return nil
}

func (s *productMasterRepoStub) GetMasterProductCounts(ctx context.Context) ([]productmaster.MasterProductCount, error) {
	return []productmaster.MasterProductCount{}, nil
}

func (s *productMasterRepoStub) UpdateMasterStatistics(ctx context.Context, masterID int64, confidence float64, matchCount int, updatedAt time.Time) (int64, error) {
	return 0, nil
}

func noopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}
