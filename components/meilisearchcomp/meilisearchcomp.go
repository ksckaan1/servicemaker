package meilisearchcomp

import (
	"context"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliSearch struct {
	sm meilisearch.ServiceManager

	Host   string `env:"MS_HOST"`
	APIKey string `env:"MS_API_KEY"`
}

func (m *MeiliSearch) Init(ctx context.Context) error {
	opts := []meilisearch.Option{}

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
