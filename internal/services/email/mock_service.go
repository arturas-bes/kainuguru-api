package email

import (
	"context"
	"fmt"

	"github.com/kainuguru/kainuguru-api/internal/models"
)

// mockEmailService provides a mock implementation for development
type mockEmailService struct{}

// NewMockService creates a new mock email service
func NewMockService() Service {
	return &mockEmailService{}
}

func (m *mockEmailService) SendVerificationEmail(ctx context.Context, user *models.User, token string) error {
	fmt.Printf("📧 MOCK EMAIL: Verification email for %s\n", user.Email)
	fmt.Printf("🔗 Verification URL: /verify-email?token=%s\n", token)
	return nil
}

func (m *mockEmailService) SendPasswordResetEmail(ctx context.Context, user *models.User, token string) error {
	fmt.Printf("📧 MOCK EMAIL: Password reset email for %s\n", user.Email)
	fmt.Printf("🔗 Reset URL: /reset-password?token=%s\n", token)
	return nil
}

func (m *mockEmailService) SendWelcomeEmail(ctx context.Context, user *models.User) error {
	fmt.Printf("📧 MOCK EMAIL: Welcome email for %s\n", user.Email)
	return nil
}

func (m *mockEmailService) SendPasswordChangedEmail(ctx context.Context, user *models.User) error {
	fmt.Printf("📧 MOCK EMAIL: Password changed notification for %s\n", user.Email)
	return nil
}

func (m *mockEmailService) SendLoginAlertEmail(ctx context.Context, user *models.User, session *models.UserSession) error {
	fmt.Printf("📧 MOCK EMAIL: Login alert for %s from %s\n", user.Email, session.GetLocationDescription())
	return nil
}
