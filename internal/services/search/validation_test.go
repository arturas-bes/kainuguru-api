package search

import (
	"testing"
)

func TestValidateSearchRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *SearchRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &SearchRequest{
				Query:  "milk",
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty query",
			req: &SearchRequest{
				Query:  "",
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "query too long",
			req: &SearchRequest{
				Query:  string(make([]byte, 256)),
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "negative min price",
			req: &SearchRequest{
				Query:    "milk",
				MinPrice: ptrFloat64(-1.0),
				Limit:    10,
				Offset:   0,
			},
			wantErr: true,
		},
		{
			name: "invalid price range",
			req: &SearchRequest{
				Query:    "milk",
				MinPrice: ptrFloat64(10.0),
				MaxPrice: ptrFloat64(5.0),
				Limit:    10,
				Offset:   0,
			},
			wantErr: true,
		},
		{
			name: "invalid limit",
			req: &SearchRequest{
				Query:  "milk",
				Limit:  0,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "limit too high",
			req: &SearchRequest{
				Query:  "milk",
				Limit:  101,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "negative offset",
			req: &SearchRequest{
				Query:  "milk",
				Limit:  10,
				Offset: -1,
			},
			wantErr: true,
		},
		{
			name: "too many store IDs",
			req: &SearchRequest{
				Query:    "milk",
				StoreIDs: make([]int, 51),
				Limit:    10,
				Offset:   0,
			},
			wantErr: true,
		},
		{
			name: "invalid store ID",
			req: &SearchRequest{
				Query:    "milk",
				StoreIDs: []int{0},
				Limit:    10,
				Offset:   0,
			},
			wantErr: true,
		},
		{
			name: "malicious query",
			req: &SearchRequest{
				Query:  "milk'; DROP TABLE products; --",
				Limit:  10,
				Offset: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSearchRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSuggestionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *SuggestionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &SuggestionRequest{
				PartialQuery: "mil",
				Limit:        5,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "empty partial query",
			req: &SuggestionRequest{
				PartialQuery: "",
				Limit:        5,
			},
			wantErr: true,
		},
		{
			name: "partial query too long",
			req: &SuggestionRequest{
				PartialQuery: string(make([]byte, 101)),
				Limit:        5,
			},
			wantErr: true,
		},
		{
			name: "invalid limit",
			req: &SuggestionRequest{
				PartialQuery: "mil",
				Limit:        0,
			},
			wantErr: true,
		},
		{
			name: "limit too high",
			req: &SuggestionRequest{
				PartialQuery: "mil",
				Limit:        21,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSuggestionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSuggestionRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "normal query",
			query: "milk bread",
			want:  "milk bread",
		},
		{
			name:  "query with extra spaces",
			query: "  milk   bread  ",
			want:  "milk bread",
		},
		{
			name:  "query with tabs and newlines",
			query: "milk\tbread\n",
			want:  "milk bread",
		},
		{
			name:  "query with control characters",
			query: "milk\x01bread\x02",
			want:  "milkbread",
		},
		{
			name:  "empty query",
			query: "",
			want:  "",
		},
		{
			name:  "only whitespace",
			query: "   \t\n  ",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeQuery(tt.query)
			if got != tt.want {
				t.Errorf("SanitizeQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSearchQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "lithuanian diacritics",
			query: "Pienas ąčęėįšųūž",
			want:  "pienas aceeisuuz",
		},
		{
			name:  "mixed case",
			query: "MILK Bread",
			want:  "milk bread",
		},
		{
			name:  "with extra spaces",
			query: "  MILK   BREAD  ",
			want:  "milk bread",
		},
		{
			name:  "empty query",
			query: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSearchQuery(tt.query)
			if got != tt.want {
				t.Errorf("NormalizeSearchQuery() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Helper function to create float64 pointer
func ptrFloat64(f float64) *float64 {
	return &f
}
