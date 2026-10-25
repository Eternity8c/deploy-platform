package features_application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	core_repo_postgres "github.com/Eternity8c/deploy-platform/internal/core/repository/postgres"
)

type Store struct {
	pool core_repo_postgres.Pool
}

func NewStore(p core_repo_postgres.Pool) *Store {
	return &Store{
		pool: p,
	}
}

type CreateResult struct {
	ID        int64     `json:"id"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Store) Create(ctx context.Context, app Application) (CreateResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.pool.OpTimeout())
	defer cancel()

	var createResponse CreateResult

	err := s.pool.QueryRow(
		ctx,
		`
		INSERT INTO applications (tenant, name, spec)
		VALUES ($1,$2,$3)
		RETURNING id, version, created_at, updated_at
		`,
		app.Spec.Tenant,
		app.Metadata.Name,
		app.Spec,
	).Scan(
		&createResponse.ID,
		&createResponse.Version,
		&createResponse.CreatedAt,
		&createResponse.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return CreateResult{}, ErrAlreadyExists
		}
		return CreateResult{}, fmt.Errorf("create application: %w", err)
	}

	return createResponse, nil
}
