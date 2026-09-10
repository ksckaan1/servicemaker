package pgcomp

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/ksckaan1/logger"
)

func (p *Postgres) MigratePostgres(ctx context.Context, fsys fs.FS, path string) error {
	efs, err := iofs.New(fsys, path)
	if err != nil {
		return fmt.Errorf("iofs.New: %w", err)
	}

	migrator, err := migrate.NewWithSourceInstance("iofs", efs, p.DBURL)
	if err != nil {
		return fmt.Errorf("migrate.NewWithSourceInstance: %w", err)
	}

	err = migrator.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		return fmt.Errorf("migrator.Up: %w", err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("migrator.Version: %w", err)
	}

	logger.Default.Info(
		ctx, "database migrated",
		"version", version,
		"dirty", dirty,
	)

	return nil
}
