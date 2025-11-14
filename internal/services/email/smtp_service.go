package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/jordan-wright/email"
	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
	"github.com/kainuguru/kainuguru-api/internal/models"
)

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// smtpEmailService implements Service using SMTP
type smtpEmailService struct {
	config    *SMTPConfig
	templates map[string]*template.Template
	logger    *slog.Logger
}

// NewSMTPService creates a new SMTP-based email service
func NewSMTPService(config *SMTPConfig) (Service, error) {
	service := &smtpEmailService{
		config:    config,
		templates: make(map[string]*template.Template),
		logger:    slog.Default().With("service", "email"),
	}

	// Load email templates
	if err := service.loadTemplates(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to load email templates")
	}

	return service, nil
}

// loadTemplates loads all email templates
func (s *smtpEmailService) loadTemplates() error {
	// Define templates inline for simplicity
	// In production, these could be loaded from files
	
	s.templates["verification"] = template.Must(template.New("verification").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; padding: 12px 30px; background: #4CAF50; color: white !important; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🍎 Verify Your Email</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>Welcome to Kainuguru! Please verify your email address to get started.</p>
            <p style="text-align: center;">
                <a href="{{.VerificationURL}}" class="button">Verify Email Address</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; font-size: 12px;">{{.VerificationURL}}</p>
            <p>This link will expire in 24 hours.</p>
            <p>If you didn't create a Kainuguru account, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>© {{.Year}} Kainuguru - Lithuanian Grocery Price Comparison</p>
        </div>
    </div>
</body>
</html>
`))

	s.templates["password_reset"] = template.Must(template.New("password_reset").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #FF9800; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; padding: 12px 30px; background: #FF9800; color: white !important; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 10px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔒 Reset Your Password</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>We received a request to reset your password for your Kainuguru account.</p>
            <p style="text-align: center;">
                <a href="{{.ResetURL}}" class="button">Reset Password</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; font-size: 12px;">{{.ResetURL}}</p>
            <p>This link will expire in 1 hour.</p>
            <div class="warning">
                <strong>⚠️ Security Notice:</strong> If you didn't request a password reset, please ignore this email. Your password will remain unchanged.
            </div>
        </div>
        <div class="footer">
            <p>© {{.Year}} Kainuguru</p>
        </div>
    </div>
</body>
</html>
`))

	s.templates["welcome"] = template.Must(template.New("welcome").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2196F3; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .feature { margin: 15px 0; padding-left: 25px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Welcome to Kainuguru!</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>Your email has been verified! You're all set to start saving money on groceries.</p>
            <h3>What you can do with Kainuguru:</h3>
            <div class="feature">🔍 Search products across all Lithuanian stores</div>
            <div class="feature">💰 Compare prices and find the best deals</div>
            <div class="feature">🛒 Create and manage shopping lists</div>
            <div class="feature">📊 Track price history and trends</div>
            <div class="feature">🔔 Get notified when new flyers are published</div>
            <p>Happy shopping!</p>
        </div>
        <div class="footer">
            <p>© {{.Year}} Kainuguru</p>
        </div>
    </div>
</body>
</html>
`))

	s.templates["password_changed"] = template.Must(template.New("password_changed").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #4CAF50; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
        .warning { background: #ffebee; border-left: 4px solid #f44336; padding: 10px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>✅ Password Changed</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>Your Kainuguru password has been successfully changed.</p>
            <p><strong>Time:</strong> {{.Timestamp}}</p>
            <div class="warning">
                <strong>⚠️ Didn't change your password?</strong><br>
                If you didn't make this change, please contact support immediately.
            </div>
        </div>
        <div class="footer">
            <p>© {{.Year}} Kainuguru</p>
        </div>
    </div>
</body>
</html>
`))

	s.templates["login_alert"] = template.Must(template.New("login_alert").Parse(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #9C27B0; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .info-box { background: white; padding: 15px; border-radius: 5px; margin: 15px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔐 New Login Detected</h1>
        </div>
        <div class="content">
            <p>Hi {{.Name}},</p>
            <p>We detected a new login to your Kainuguru account:</p>
            <div class="info-box">
                <p><strong>Time:</strong> {{.Timestamp}}</p>
                <p><strong>Location:</strong> {{.Location}}</p>
                <p><strong>Device:</strong> {{.Device}}</p>
                <p><strong>IP Address:</strong> {{.IPAddress}}</p>
            </div>
            <p>If this was you, you can safely ignore this email.</p>
            <p>If you don't recognize this activity, please secure your account immediately.</p>
        </div>
        <div class="footer">
            <p>© {{.Year}} Kainuguru</p>
        </div>
    </div>
</body>
</html>
`))

	return nil
}

// SendVerificationEmail sends email verification
func (s *smtpEmailService) SendVerificationEmail(ctx context.Context, user *models.User, token string) error {
	data := map[string]interface{}{
		"Name":            getUserDisplayName(user),
		"VerificationURL": fmt.Sprintf("%s/verify-email?token=%s", s.getBaseURL(), token),
		"Year":            time.Now().Year(),
	}

	subject := "Verify Your Kainuguru Account"
	return s.sendTemplateEmail(ctx, user.Email, subject, "verification", data)
}

// SendPasswordResetEmail sends password reset email
func (s *smtpEmailService) SendPasswordResetEmail(ctx context.Context, user *models.User, token string) error {
	data := map[string]interface{}{
		"Name":     getUserDisplayName(user),
		"ResetURL": fmt.Sprintf("%s/reset-password?token=%s", s.getBaseURL(), token),
		"Year":     time.Now().Year(),
	}

	subject := "Reset Your Kainuguru Password"
	return s.sendTemplateEmail(ctx, user.Email, subject, "password_reset", data)
}

// SendWelcomeEmail sends welcome email after verification
func (s *smtpEmailService) SendWelcomeEmail(ctx context.Context, user *models.User) error {
	data := map[string]interface{}{
		"Name": getUserDisplayName(user),
		"Year": time.Now().Year(),
	}

	subject := "Welcome to Kainuguru! 🎉"
	return s.sendTemplateEmail(ctx, user.Email, subject, "welcome", data)
}

// SendPasswordChangedEmail sends notification when password is changed
func (s *smtpEmailService) SendPasswordChangedEmail(ctx context.Context, user *models.User) error {
	data := map[string]interface{}{
		"Name":      getUserDisplayName(user),
		"Timestamp": time.Now().Format("January 2, 2006 at 3:04 PM MST"),
		"Year":      time.Now().Year(),
	}

	subject := "Your Kainuguru Password Has Been Changed"
	return s.sendTemplateEmail(ctx, user.Email, subject, "password_changed", data)
}

// SendLoginAlertEmail sends alert for new login
func (s *smtpEmailService) SendLoginAlertEmail(ctx context.Context, user *models.User, session *models.UserSession) error {
	data := map[string]interface{}{
		"Name":      getUserDisplayName(user),
		"Timestamp": session.CreatedAt.Format("January 2, 2006 at 3:04 PM MST"),
		"Location":  session.GetLocationDescription(),
		"Device":    session.GetDeviceDescription(),
		"IPAddress": session.IPAddress,
		"Year":      time.Now().Year(),
	}

	subject := "New Login to Your Kainuguru Account"
	return s.sendTemplateEmail(ctx, user.Email, subject, "login_alert", data)
}

// sendTemplateEmail sends an email using a template
func (s *smtpEmailService) sendTemplateEmail(ctx context.Context, to, subject, templateName string, data map[string]interface{}) error {
	tmpl, exists := s.templates[templateName]
	if !exists {
		return apperrors.NotFound(fmt.Sprintf("template %s not found", templateName))
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to execute template")
	}

	return s.sendEmail(ctx, to, subject, body.String())
}

// sendEmail sends an email via SMTP
func (s *smtpEmailService) sendEmail(ctx context.Context, to, subject, htmlBody string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.From)
	e.To = []string{to}
	e.Subject = subject
	e.HTML = []byte(htmlBody)
	e.Headers = textproto.MIMEHeader{}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	var auth smtp.Auth
	
	// Only use auth if username is provided
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	s.logger.Info("sending email",
		slog.String("to", to),
		slog.String("subject", subject),
		slog.String("smtp_host", s.config.Host),
	)

	// Send with timeout
	done := make(chan error, 1)
	go func() {
		done <- e.Send(addr, auth)
	}()

	select {
	case err := <-done:
		if err != nil {
			s.logger.Error("failed to send email",
				slog.String("to", to),
				slog.String("error", err.Error()),
			)
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to send email")
		}
		s.logger.Info("email sent successfully", slog.String("to", to))
		return nil
	case <-time.After(30 * time.Second):
		return apperrors.Internal("email send timeout after 30 seconds")
	case <-ctx.Done():
		return apperrors.Wrap(ctx.Err(), apperrors.ErrorTypeInternal, "email send cancelled")
	}
}

// getBaseURL returns the base URL for the application
func (s *smtpEmailService) getBaseURL() string {
	// In production, this should come from configuration
	return "https://kainuguru.lt"
}

// getUserDisplayName returns a display name for the user
func getUserDisplayName(user *models.User) string {
	if user.FullName != nil && *user.FullName != "" {
		return *user.FullName
	}
	return user.Email
}
