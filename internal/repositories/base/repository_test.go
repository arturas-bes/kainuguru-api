package base

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
	"strings"
)

type testEntity struct {
	bun.BaseModel `bun:"table:test_entities,alias:te"`

	ID        int       `bun:"id,pk,autoincrement"`
	Name      string    `bun:"name,notnull"`
	IsActive  bool      `bun:"is_active,notnull"`
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp"`
}

func TestRepository_GetByID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, cleanup := newTestRepository(t)
	defer cleanup()

	entity := &testEntity{Name: "store-front", IsActive: true}
	require.NoError(t, repo.Create(ctx, entity))

	fetched, err := repo.GetByID(ctx, entity.ID)
	require.NoError(t, err)
	require.Equal(t, entity.Name, fetched.Name)
	require.True(t, fetched.IsActive)
}

func TestRepository_GetByIDsEmptyInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, cleanup := newTestRepository(t)
	defer cleanup()

	results, err := repo.GetByIDs(ctx, []int{})
	require.NoError(t, err)
	require.Empty(t, results)
}

func TestRepository_GetAllWithPagination(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, cleanup := newTestRepository(t)
	defer cleanup()

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(ctx, &testEntity{
			Name:     time.Now().Add(time.Duration(i) * time.Second).Format(time.RFC3339),
			IsActive: true,
		}))
	}

	results, err := repo.GetAll(ctx, WithLimit[testEntity](2), WithOffset[testEntity](2), WithOrderClause[testEntity]("te.id ASC"))
	require.NoError(t, err)
	require.Len(t, results, 2)
	require.Greater(t, results[0].ID, 0)
}

func TestRepository_Count(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, cleanup := newTestRepository(t)
	defer cleanup()

	require.NoError(t, repo.Create(ctx, &testEntity{Name: "a", IsActive: true}))
	require.NoError(t, repo.Create(ctx, &testEntity{Name: "b", IsActive: false}))

	count, err := repo.Count(ctx, WithQuery[testEntity](func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("te.is_active = ?", true)
	}))
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestRepository_CreateManyAndDelete(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, cleanup := newTestRepository(t)
	defer cleanup()

	entities := []*testEntity{
		{Name: "foo", IsActive: true},
		{Name: "bar", IsActive: true},
	}
	require.NoError(t, repo.CreateMany(ctx, entities))

	results, err := repo.GetAll(ctx, WithOrderClause[testEntity]("te.id ASC"))
	require.NoError(t, err)
	require.Len(t, results, 2)

	require.NoError(t, repo.DeleteByID(ctx, results[0].ID))
	remaining, err := repo.GetAll(ctx)
	require.NoError(t, err)
	require.Len(t, remaining, 1)
}

func TestRepository_Update(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo, cleanup := newTestRepository(t)
	defer cleanup()

	entity := &testEntity{Name: "before", IsActive: true}
	require.NoError(t, repo.Create(ctx, entity))
	entity.Name = "after"

	require.NoError(t, repo.Update(ctx, entity))

	updated, err := repo.GetByID(ctx, entity.ID)
	require.NoError(t, err)
	require.Equal(t, "after", updated.Name)
}

func newTestRepository(t *testing.T) (*Repository[testEntity], func()) {
	t.Helper()

	dsn := fmt.Sprintf("file:baserepo_%s?mode=memory&cache=shared", sanitizeName(t.Name()))
	sqlDB, err := sql.Open(sqliteshim.ShimName, dsn)
	require.NoError(t, err)

	db := bun.NewDB(sqlDB, sqlitedialect.New())
	ctx := context.Background()

	_, err = db.NewCreateTable().
		Model((*testEntity)(nil)).
		IfNotExists().
		Exec(ctx)
	require.NoError(t, err)

	repo := NewRepository[testEntity](db, "id")

	cleanup := func() {
		_ = db.Close()
		_ = sqlDB.Close()
	}
	return repo, cleanup
}

func sanitizeName(s string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", " ", "_")
	return replacer.Replace(s)
}
