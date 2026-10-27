package meilisearchcomp

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type MeiliSearch struct {
	sm meilisearch.ServiceManager

	Host   string `env:"MS_HOST,required"`
	APIKey string `env:"MS_API_KEY"`
}

func (m *MeiliSearch) Init(ctx context.Context) error {
	opts := []meilisearch.Option{
		// Traces and measures every request made to Meilisearch. Passing a
		// custom client replaces the library's own transport, so the defaults
		// it uses are mirrored here.
		meilisearch.WithCustomClient(&http.Client{
			Transport: otelhttp.NewTransport(&http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}),
		}),
	}

	if m.APIKey != "" {
		opts = append(opts, meilisearch.WithAPIKey(m.APIKey))
	}

	m.sm = meilisearch.New(m.Host, opts...)

	return nil
}

func (m *MeiliSearch) Close(ctx context.Context) error {
	m.sm.Close()
	return nil
}

func (m *MeiliSearch) HealthCheck(ctx context.Context) error {
	_, err := m.sm.HealthWithContext(ctx)
	if err != nil {
		return fmt.Errorf("sm.HealthWithContext: %w", err)
	}
	return nil
}

func (m *MeiliSearch) ServiceManager() meilisearch.ServiceManager {
	return m.sm
}
