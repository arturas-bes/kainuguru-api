package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// UserStorePreference represents a user's preferred store
type UserStorePreference struct {
	bun.BaseModel `bun:"table:user_store_preferences,alias:usp"`

	ID        int       `bun:"id,pk,autoincrement" json:"id"`
	UserID    uuid.UUID `bun:"user_id,notnull,type:uuid" json:"userId"`
	StoreID   int       `bun:"store_id,notnull" json:"storeId"`
	CreatedAt time.Time `bun:"created_at,default:now()" json:"createdAt"`

	// Relations
	User  *User  `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Store *Store `bun:"rel:belongs-to,join:store_id=id" json:"store,omitempty"`
}

// TableName returns the table name for Bun
func (usp *UserStorePreference) TableName() string {
	return "user_store_preferences"
}
