package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ProductMasterMatch struct {
	bun.BaseModel `bun:"table:product_master_matches,alias:pmm"`

	ID              int64      `bun:"id,pk,autoincrement" json:"id"`
	ProductID       int64      `bun:"product_id,notnull" json:"product_id"`
	ProductMasterID int64      `bun:"product_master_id,notnull" json:"product_master_id"`
	Confidence      float64    `bun:"confidence,notnull" json:"confidence"`
	MatchType       string     `bun:"match_type,notnull" json:"match_type"`
	MatchScore      *float64   `bun:"match_score" json:"match_score"`
	MatchedFields   any        `bun:"matched_fields,type:jsonb" json:"matched_fields"`
	ReviewStatus    string     `bun:"review_status,default:'pending'" json:"review_status"`
	ReviewedBy      *string    `bun:"reviewed_by,type:uuid" json:"reviewed_by"`
	ReviewedAt      *time.Time `bun:"reviewed_at" json:"reviewed_at"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull,default:now()" json:"updated_at"`

	Product *Product       `bun:"rel:belongs-to,join:product_id=id" json:"product,omitempty"`
	Master  *ProductMaster `bun:"rel:belongs-to,join:product_master_id=id" json:"master,omitempty"`
}

func (pmm *ProductMasterMatch) TableName() string {
	return "product_master_matches"
}
