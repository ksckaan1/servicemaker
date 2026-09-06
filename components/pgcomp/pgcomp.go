package pgcomp

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	DBURL string `env:"PG_DB_URL"`
	db    *pgxpool.Pool
}

func (p *Postgres) Init(ctx context.Context) error {
	var err error

	p.db, err = pgxpool.New(ctx, p.DBURL)
	if err != nil {
		return fmt.Errorf("pgxpool.New: %w", err)
	}

	return nil
}

func (p *Postgres) Close(ctx context.Context) error {
	p.db.Close()

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
