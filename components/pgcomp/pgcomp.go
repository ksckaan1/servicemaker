package pgcomp

import (
	"context"
	"fmt"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksckaan1/logger"
)

type Postgres struct {
	DBURL string `env:"PG_DB_URL,required"`
	db    *pgxpool.Pool
}

func (p *Postgres) Init(ctx context.Context) error {
	cfg, err := pgxpool.ParseConfig(p.DBURL)
	if err != nil {
		return fmt.Errorf("pgxpool.ParseConfig: %w", err)
	}

	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	p.db, err = pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	err = otelpgx.RecordStats(p.db)
	if err != nil {
		return fmt.Errorf("otelpgx.RecordStats: %w", err)
	}

	logger.Default.Info(ctx, "postgres client initialized")

	return nil
}

func (p *Postgres) Close(ctx context.Context) error {
	p.db.Close()
	logger.Default.Info(ctx, "postgres client closed")

	return nil
}

func (p *Postgres) HealthCheck(ctx context.Context) error {
	err := p.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("pg.Ping: %w", err)
	}

	return nil
}

func (p *Postgres) DB() *pgxpool.Pool {
	return p.db
}
