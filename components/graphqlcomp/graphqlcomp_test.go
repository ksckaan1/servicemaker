package graphqlcomp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Verifies that GQL_OTEL_SKIP_PATHS keeps the given paths out of traces.
func TestOtelSkipPaths(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))

	g := &GraphQL{Addr: ":0", OtelSkipPaths: []string{"/health"}}
	if err := g.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	g.router.Get("/health", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	g.router.Post("/graphql", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })

	for _, path := range []string{"/health", "/graphql"} {
		method := http.MethodGet
		if path == "/graphql" {
			method = http.MethodPost
		}

		req := httptest.NewRequest(method, path, nil)
		resp, err := g.router.Test(req)
		if err != nil {
			t.Fatalf("router.Test(%s): %v", path, err)
		}
		resp.Body.Close()
	}

	var names []string
	for _, s := range rec.Ended() {
		names = append(names, s.Name())
	}

	if len(names) != 1 {
		t.Fatalf("expected exactly 1 span for /graphql, got %v", names)
	}
}
