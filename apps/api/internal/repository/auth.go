package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
)

type AuthRepository interface {
	RevokeRefreshToken(ctx context.Context, token string) error
	GetRefreshTokenByHash(ctx context.Context, token string) (db.CoreRefreshToken, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.CoreUser, error)
	CreateRefreshToken(ctx context.Context, params db.CreateRefreshTokenParams) (db.CoreRefreshToken, error)
}

type AuthTxRepository interface {
	WithTx(tx pgx.Tx) AuthTxRepository
	GetUserByEmail(ctx context.Context, email string) (db.CoreUser, error)
	CreateUser(ctx context.Context, params db.CreateUserParams) (db.CoreUser, error)
	CreateAccount(ctx context.Context, params db.CreateAccountParams) (db.CoreAccount, error)
}

type authTxRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewAuthTxRepository(queries *db.Queries, pool *pgxpool.Pool) AuthTxRepository {
	return &authTxRepository{
		queries: queries,
		pool:    pool,
	}
}

func (r *authTxRepository) WithTx(tx pgx.Tx) AuthTxRepository {
	return &authTxRepository{
		queries: r.queries.WithTx(tx),
		pool:    r.pool,
	}
}

func (r *authTxRepository) GetUserByEmail(ctx context.Context, email string) (db.CoreUser, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *authTxRepository) CreateUser(ctx context.Context, params db.CreateUserParams) (db.CoreUser, error) {
	return r.queries.CreateUser(ctx, params)
}

func (r *authTxRepository) CreateAccount(ctx context.Context, params db.CreateAccountParams) (db.CoreAccount, error) {
	return r.queries.CreateAccount(ctx, params)
}
