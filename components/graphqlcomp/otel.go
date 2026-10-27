package graphqlcomp

import (
	"context"
	"sync"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// otelExtension adds a span per GraphQL operation. The HTTP middleware only
// reports the transport request, which carries no operation name or type.
type otelExtension struct{}

func (otelExtension) ExtensionName() string { return "OtelTracing" }

func (otelExtension) Validate(graphql.ExecutableSchema) error { return nil }

func (otelExtension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	oc := graphql.GetOperationContext(ctx)

	name := "anonymous"
	opType := "unknown"
	if oc.Operation != nil {
		if oc.Operation.Name != "" {
			name = oc.Operation.Name
		}
		opType = string(oc.Operation.Operation)
	}

	ctx, span := otel.Tracer("graphqlcomp").Start(ctx, "graphql.operation",
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("graphql.operation.name", name),
			attribute.String("graphql.operation.type", opType),
		),
	)

	// A subscription invokes the handler once per event, so the span only ends
	// when the stream closes.
	subscription := oc.Operation != nil && oc.Operation.Operation == ast.Subscription
	responses := next(ctx)
	var end sync.Once

	return func(ctx context.Context) *graphql.Response {
		resp := responses(ctx)

		if resp != nil {
			for _, gqlErr := range resp.Errors {
				span.RecordError(gqlErr)
			}
			if len(resp.Errors) > 0 {
				span.SetStatus(codes.Error, resp.Errors[0].Message)
			}
		}

		if !subscription || resp == nil {
			end.Do(func() { span.End() })
		}

		return resp
	}
}
