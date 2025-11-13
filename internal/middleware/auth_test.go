package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kainuguru/kainuguru-api/internal/models"
	"github.com/kainuguru/kainuguru-api/internal/services/auth"
)

func TestNewAuthMiddleware_RequiredMissingTokenReturnsUnauthorized(t *testing.T) {
	app := fiber.New()
	jwtStub := &jwtServiceStub{}
	sessionStub := &sessionServiceStub{}
	app.Use(NewAuthMiddleware(AuthMiddlewareConfig{
		Required:       true,
		JWTService:     jwtStub,
		SessionService: sessionStub,
	}))

	called := false
	app.Get("/", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", resp.StatusCode)
	}
	if called {
		t.Fatalf("handler should not run when auth fails")
	}
}

func TestNewAuthMiddleware_OptionalMissingTokenProceeds(t *testing.T) {
	app := fiber.New()
	jwtStub := &jwtServiceStub{
		validateCalled: func(token string) {
			t.Fatalf("ValidateAccessToken should not be called without token")
		},
	}
	sessionStub := &sessionServiceStub{}
	app.Use(NewAuthMiddleware(AuthMiddlewareConfig{
		Required:       false,
		JWTService:     jwtStub,
		SessionService: sessionStub,
	}))

	called := false
	app.Get("/", func(c *fiber.Ctx) error {
		called = true
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for optional auth, got %d", resp.StatusCode)
	}
	if !called {
		t.Fatalf("handler should run when auth optional")
	}
}

func TestNewAuthMiddleware_SetsContextOnSuccess(t *testing.T) {
	userID := uuid.New()
	sessionID := uuid.New()

	app := fiber.New()
	jwtStub := &jwtServiceStub{
		validateAccessTokenFunc: func(token string) (*auth.TokenClaims, error) {
			if token != "good-token" {
				return nil, errors.New("unexpected token")
			}
			return &auth.TokenClaims{
				UserID:    userID,
				SessionID: sessionID,
				ExpiresAt: time.Now().Add(time.Hour),
			}, nil
		},
	}
	sessionStub := &sessionServiceStub{
		validateSessionFunc: func(ctx context.Context, id uuid.UUID) (*models.UserSession, error) {
			if id != sessionID {
				return nil, errors.New("unexpected session id")
			}
			return &models.UserSession{ID: id}, nil
		},
	}

	app.Use(NewAuthMiddleware(AuthMiddlewareConfig{
		Required:       true,
		JWTService:     jwtStub,
		SessionService: sessionStub,
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		gotUser, ok := GetUserFromContext(ctx)
		if !ok || gotUser != userID {
			return fiber.NewError(fiber.StatusInternalServerError, "user ID not set on context")
		}
		gotSession, ok := GetSessionFromContext(ctx)
		if !ok || gotSession != sessionID {
			return fiber.NewError(fiber.StatusInternalServerError, "session ID not set on context")
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 when auth succeeds, got %d", resp.StatusCode)
	}
}

// --- stubs ---

type jwtServiceStub struct {
	validateAccessTokenFunc func(token string) (*auth.TokenClaims, error)
	validateCalled          func(token string)
}

func (s *jwtServiceStub) GenerateTokenPair(userID uuid.UUID, sessionID uuid.UUID) (*auth.TokenPair, error) {
	return nil, errors.New("not implemented")
}

func (s *jwtServiceStub) ValidateAccessToken(token string) (*auth.TokenClaims, error) {
	if s.validateCalled != nil {
		s.validateCalled(token)
	}
	if s.validateAccessTokenFunc != nil {
		return s.validateAccessTokenFunc(token)
	}
	return nil, errors.New("not implemented")
}

func (s *jwtServiceStub) ValidateRefreshToken(token string) (*auth.TokenClaims, error) {
	return nil, errors.New("not implemented")
}

func (s *jwtServiceStub) GetTokenHash(token string) string {
	return ""
}

func (s *jwtServiceStub) ExtractClaims(token string) (*auth.TokenClaims, error) {
	return nil, errors.New("not implemented")
}

type sessionServiceStub struct {
	validateSessionFunc func(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error)
}

func (s *sessionServiceStub) CreateSession(ctx context.Context, input *models.SessionCreateInput) (*models.UserSession, error) {
	return nil, errors.New("not implemented")
}

func (s *sessionServiceStub) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error) {
	return nil, errors.New("not implemented")
}

func (s *sessionServiceStub) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.UserSession, error) {
	return nil, errors.New("not implemented")
}

func (s *sessionServiceStub) ValidateSession(ctx context.Context, sessionID uuid.UUID) (*models.UserSession, error) {
	if s.validateSessionFunc != nil {
		return s.validateSessionFunc(ctx, sessionID)
	}
	return nil, errors.New("not implemented")
}

func (s *sessionServiceStub) UpdateSessionActivity(ctx context.Context, sessionID uuid.UUID) error {
	return errors.New("not implemented")
}

func (s *sessionServiceStub) InvalidateSession(ctx context.Context, sessionID uuid.UUID) error {
	return errors.New("not implemented")
}

func (s *sessionServiceStub) InvalidateUserSessions(ctx context.Context, userID uuid.UUID) error {
	return errors.New("not implemented")
}

func (s *sessionServiceStub) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return 0, errors.New("not implemented")
}

func (s *sessionServiceStub) GetUserSessions(ctx context.Context, userID uuid.UUID, filters *models.SessionFilters) ([]*models.UserSession, error) {
	return nil, errors.New("not implemented")
}
