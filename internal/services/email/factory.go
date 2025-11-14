package email

import (
	"fmt"
	"strings"

	"github.com/kainuguru/kainuguru-api/internal/config"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

// NewEmailServiceFromConfig creates an email service based on configuration
func NewEmailServiceFromConfig(cfg *config.Config) (Service, error) {
	provider := strings.ToLower(cfg.Email.Provider)

	switch provider {
	case "smtp":
		smtpConfig := &SMTPConfig{
			Host:     cfg.Email.SMTP.Host,
			Port:     cfg.Email.SMTP.Port,
			Username: cfg.Email.SMTP.Username,
			Password: cfg.Email.SMTP.Password,
			From:     cfg.Email.FromEmail,
			FromName: cfg.Email.FromName,
		}

		if smtpConfig.Host == "" {
			return nil, apperrors.Validation("SMTP host is required when using SMTP provider")
		}

		if smtpConfig.From == "" {
			return nil, apperrors.Validation("from email is required when using SMTP provider")
		}

		return NewSMTPService(smtpConfig)

	case "mock", "":
		// Use mock service in development
		return NewMockService(), nil

	default:
		return nil, apperrors.NotFound(fmt.Sprintf("unknown email provider: %s (supported: smtp, mock)", provider))
	}
}
