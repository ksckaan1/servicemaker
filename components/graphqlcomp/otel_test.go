package graphqlcomp

import (
	"context"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Verifies that an operation span with its name and type is produced, since
// the HTTP middleware alone carries no GraphQL semantics.
func TestOtelExtensionOperationSpan(t *testing.T) {
	rec := tracetest.NewSpanRecorder()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))

	oc := &graphql.OperationContext{
		Operation: &ast.OperationDefinition{Name: "GetUser", Operation: ast.Query},
	}
	ctx := graphql.WithOperationContext(context.Background(), oc)

	next := func(context.Context) graphql.ResponseHandler {
		return func(context.Context) *graphql.Response {
			return &graphql.Response{Data: []byte(`{"ok":true}`)}
		}
	}

	responses := otelExtension{}.InterceptOperation(ctx, next)
	if resp := responses(ctx); resp == nil {
		t.Fatal("no response")
	}

	spans := rec.Ended()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name() != "graphql.operation" {
		t.Fatalf("span name = %q", span.Name())
	}

	attrs := map[string]string{}
	for _, kv := range span.Attributes() {
		attrs[string(kv.Key)] = kv.Value.AsString()
	}
	if attrs["graphql.operation.name"] != "GetUser" {
		t.Fatalf("operation name = %q", attrs["graphql.operation.name"])
	}
	if attrs["graphql.operation.type"] != "query" {
		t.Fatalf("operation type = %q", attrs["graphql.operation.type"])
	}
}
