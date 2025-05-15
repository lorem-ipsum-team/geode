package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	postgres_mig "github.com/lorem-ipsum-team/geode/db/postgres"
	"github.com/lorem-ipsum-team/geode/logger"
)

type Repo struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewRepo(ctx context.Context, log *slog.Logger, connString string) (Repo, error) {
	log = log.WithGroup("postgres_repo")
	log.Info("connect to db", slog.Any("connection string", logger.Secret(connString)))
	log.InfoContext(ctx, "running migrations")

	err := postgres_mig.Up(ctx, connString)
	if err != nil {
		return Repo{}, fmt.Errorf("failed to run migration: %w", err)
	}

	log.InfoContext(ctx, "creating pgx connection pool")

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return Repo{}, err
	}

	return Repo{
		pool: pool,
		log:  log,
	}, nil
}

func (r Repo) UpsertGeoData(ctx context.Context, userID uuid.UUID, long, lat float64) error {
	query := `INSERT INTO geo (id, location, geo_updated)
				VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), now())
				ON CONFLICT (id) DO UPDATE
				SET 
  					location = EXCLUDED.location,
  					geo_updated = EXCLUDED.geo_updated;`

	_, err := r.pool.Exec(ctx, query, userID, long, lat)
	if err != nil {
		return fmt.Errorf("failed to upsert geo data: %w", err)
	}

	return nil
}
