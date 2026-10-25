package otelcomp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
)

// Verifies that Close flushes spans that are still buffered in the batch
// processor, which is the whole point of registering this component.
func TestCloseFlushesBufferedSpans(t *testing.T) {
	restoreTransport(t)

	var got int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_SERVICE_NAME", "otelcomp-test")

	o := &Otel{}
	if err := o.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	_, span := otel.Tracer("test").Start(context.Background(), "test-span")
	span.End()

	if got != 0 {
		t.Fatalf("span exported before Close, got %d requests", got)
	}

	if err := o.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if got == 0 {
		t.Fatal("Close did not flush buffered spans")
	}
}

// Verifies that plain http.DefaultClient requests are traced and that the
// trace context header reaches the downstream service.
func TestDefaultTransportIsTraced(t *testing.T) {
	restoreTransport(t)

	var traceparent string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent = r.Header.Get("traceparent")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_SERVICE_NAME", "otelcomp-test")

	o := &Otel{}
	if err := o.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}
	defer o.Close(context.Background())

	resp, err := http.DefaultClient.Get(srv.URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	resp.Body.Close()

	if traceparent == "" {
		t.Fatal("traceparent header was not injected by the default transport")
	}
}

func restoreTransport(t *testing.T) {
	t.Helper()

	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })
}

// Verifies that documented OTEL_TRACES_SAMPLER values are honoured by the SDK.
func TestSamplerEnvIsHonoured(t *testing.T) {
	restoreTransport(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_SERVICE_NAME", "otelcomp-test")

	for _, tc := range []struct {
		sampler     string
		wantSampled bool
	}{
		{sampler: "always_off", wantSampled: false},
		{sampler: "always_on", wantSampled: true},
	} {
		t.Setenv("OTEL_TRACES_SAMPLER", tc.sampler)

		o := &Otel{}
		if err := o.Init(context.Background()); err != nil {
			t.Fatalf("Init(%s): %v", tc.sampler, err)
		}

		_, span := otel.Tracer("test").Start(context.Background(), "span")
		sampled := span.SpanContext().IsSampled()
		span.End()

		if sampled != tc.wantSampled {
			t.Fatalf("OTEL_TRACES_SAMPLER=%s: sampled=%v, want %v", tc.sampler, sampled, tc.wantSampled)
		}

		if err := o.Close(context.Background()); err != nil {
			t.Fatalf("Close: %v", err)
		}
	}
}
