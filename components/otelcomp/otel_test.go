package otelcomp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
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

// Verifies that logs emitted through the global logger provider are buffered
// and flushed to the OTLP endpoint on Close, and that they carry the resource.
func TestCloseFlushesBufferedLogs(t *testing.T) {
	restoreTransport(t)

	bodies := make(chan string, 8)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.URL.Path == "/v1/logs" {
			bodies <- string(body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_SERVICE_NAME", "otelcomp-logs-test")

	o := &Otel{}
	if err := o.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	logger := global.GetLoggerProvider().Logger("test")
	var r log.Record
	r.SetBody(attribute.StringValue("log-from-test"))
	logger.Emit(context.Background(), r)

	select {
	case <-bodies:
		t.Fatal("log exported before Close")
	default:
	}

	if err := o.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}

	select {
	case body := <-bodies:
		if !strings.Contains(body, "log-from-test") {
			t.Fatalf("exported log does not contain the body: %q", body)
		}
		if !strings.Contains(body, "otelcomp-logs-test") {
			t.Fatalf("exported log does not contain the service name: %q", body)
		}
	default:
		t.Fatal("Close did not flush buffered logs")
	}
}

func restoreTransport(t *testing.T) {
	t.Helper()

	orig := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = orig })
}

// Verifies that metrics recorded through the global meter provider are
// exported to the OTLP endpoint on Shutdown.
func TestCloseFlushesMetrics(t *testing.T) {
	restoreTransport(t)

	metricBodies := make(chan string, 8)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/metrics" {
			body, _ := io.ReadAll(r.Body)
			metricBodies <- string(body)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
	t.Setenv("OTEL_SERVICE_NAME", "otelcomp-metrics-test")

	o := &Otel{}
	if err := o.Init(context.Background()); err != nil {
		t.Fatalf("Init: %v", err)
	}

	counter, err := otel.Meter("test").Int64Counter("test.counter")
	if err != nil {
		t.Fatalf("Int64Counter: %v", err)
	}
	counter.Add(context.Background(), 7)

	select {
	case <-metricBodies:
		t.Fatal("metrics exported before Close")
	default:
	}

	if err := o.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}

	select {
	case body := <-metricBodies:
		if !strings.Contains(body, "test.counter") {
			t.Fatalf("exported metrics do not contain the instrument name: %q", body)
		}
		if !strings.Contains(body, "otelcomp-metrics-test") {
			t.Fatalf("exported metrics do not contain the service name: %q", body)
		}
	default:
		t.Fatal("Close did not flush metrics")
	}
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
