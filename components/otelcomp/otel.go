package otelcomp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
)

// Otel configures the global tracer provider. Register it last so its batch
// flush runs after the other components have closed and ended their spans.
type Otel struct {
	tp *sdktrace.TracerProvider

	// CONFIGS
	ServiceName string `env:"OTEL_SERVICE_NAME"`
}

func (o *Otel) Init(ctx context.Context) error {
	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		return fmt.Errorf("otlptracehttp.New: %w", err)
	}

	serviceName := o.ServiceName
	if serviceName == "" {
		serviceName = "unknown_service"
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(semconv.ServiceName(serviceName)),
		resource.WithFromEnv(),
		resource.WithHost(),
	)
	if err != nil {
		return fmt.Errorf("resource.New: %w", err)
	}

	o.tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(o.tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	// Traces every request made with http.DefaultClient (and any client that
	// did not set its own Transport), injecting the W3C trace context header.
	http.DefaultTransport = otelhttp.NewTransport(http.DefaultTransport)

	return nil
}

// Close flushes buffered spans. The lifecycle context is already canceled by
// the time closers run, so a fresh context is required.
func (o *Otel) Close(context.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := o.tp.Shutdown(ctx); err != nil {
		return fmt.Errorf("tracerProvider.Shutdown: %w", err)
	}

	return nil
}
