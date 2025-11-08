package models

import (
	"time"

	"github.com/uptrace/bun"
)

type ProductMasterMatch struct {
	bun.BaseModel `bun:"table:product_master_matches,alias:pmm"`

	ID              int64     `bun:"id,pk,autoincrement" json:"id"`
	ProductID       int64     `bun:"product_id,notnull" json:"product_id"`
	MasterID        int64     `bun:"master_id,notnull" json:"master_id"`
	ConfidenceScore float64   `bun:"confidence_score,notnull" json:"confidence_score"`
	MatchMethod     string    `bun:"match_method,notnull" json:"match_method"`
	MatchedAt       time.Time `bun:"matched_at,notnull,default:now()" json:"matched_at"`
	MatchedBy       string    `bun:"matched_by" json:"matched_by"`
	IsVerified      bool      `bun:"is_verified,default:false" json:"is_verified"`

	Product *Product       `bun:"rel:belongs-to,join:product_id=id" json:"product,omitempty"`
	Master  *ProductMaster `bun:"rel:belongs-to,join:master_id=id" json:"master,omitempty"`
}

func (pmm *ProductMasterMatch) TableName() string {
	return "product_master_matches"
}
