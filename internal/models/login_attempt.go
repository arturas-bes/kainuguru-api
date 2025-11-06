package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// LoginAttempt represents a login attempt record
type LoginAttempt struct {
	bun.BaseModel `bun:"table:login_attempts,alias:la"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	Email     string    `bun:"email,notnull" json:"email"`
	Success   bool      `bun:"success,notnull" json:"success"`
	IPAddress *string   `bun:"ip_address" json:"ipAddress"`
	UserAgent *string   `bun:"user_agent" json:"userAgent"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"createdAt"`
}