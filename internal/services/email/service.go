package email

import (
	"context"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// Service defines the interface for email operations
// This is separate from auth - follows enterprise separation of concerns
type Service interface {
	SendVerificationEmail(ctx context.Context, user *models.User, token string) error
	SendPasswordResetEmail(ctx context.Context, user *models.User, token string) error
	SendWelcomeEmail(ctx context.Context, user *models.User) error
	SendPasswordChangedEmail(ctx context.Context, user *models.User) error
	SendLoginAlertEmail(ctx context.Context, user *models.User, session *models.UserSession) error
}
