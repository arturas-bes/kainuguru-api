package resolvers

import "testing"

func TestNewPaginationArgsAppliesDefaults(t *testing.T) {
	p := newDefaultPagination(nil, nil)
	if p.Limit() != 20 || p.Offset() != 0 {
		t.Fatalf("expected default limit 20/offset 0, got %d/%d", p.Limit(), p.Offset())
	}
	if p.LimitWithExtra() != 21 {
		t.Fatalf("expected limit+1 to be 21, got %d", p.LimitWithExtra())
	}
	if p.HasPreviousPage() {
		t.Fatalf("expected no previous page when offset is zero")
	}
}

func TestNewPaginationArgsCapsMax(t *testing.T) {
	first := 500
	cursor := encodeCursor(10)
	p := newDefaultPagination(&first, &cursor)
	if p.Limit() != 100 {
		t.Fatalf("expected limit capped at 100, got %d", p.Limit())
	}
	if p.Offset() != 10 {
		t.Fatalf("expected offset 10 from cursor, got %d", p.Offset())
	}
	if !p.HasPreviousPage() {
		t.Fatalf("expected previous page when offset > 0")
	}
}

func TestCustomPaginationArgs(t *testing.T) {
	first := 5
	p := newPaginationArgs(&first, nil, paginationDefaults{defaultLimit: 50, maxLimit: 60})
	if p.Limit() != 5 {
		t.Fatalf("expected limit 5, got %d", p.Limit())
	}
}
