package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestGraphQLHandlerPropagatesRequestContextCancellation ensures we keep the original
// request context (or a user-provided one) so downstream code sees cancellation.
func TestGraphQLHandlerPropagatesRequestContextCancellation(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	reqCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	var observedCtx context.Context

	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/graphql" {
			c.SetUserContext(reqCtx)
		}

		if err := c.Next(); err != nil {
			return err
		}

		if c.Path() == "/graphql" {
			observedCtx = c.UserContext()
		}
		return nil
	})

	app.All("/graphql", GraphQLHandler(GraphQLConfig{}))

	body := `{"query":"{ __typename }"}`
	req := httptest.NewRequest(http.MethodPost, "/graphql", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	t.Cleanup(func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	})

	require.NotNil(t, observedCtx, "handler should store a context on the request")

	// Cancel the original request context and ensure the handler-provided context observes it.
	cancel()

	select {
	case <-observedCtx.Done():
	case <-time.After(time.Second):
		t.Fatal("downstream context did not observe cancellation")
	}
}
