package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, params db.CreateUserParams) (db.CoreUser, error)
	RevokeRefreshToken(ctx context.Context, token string) error
	GetRefreshTokenByHash(ctx context.Context, token string) (db.CoreRefreshToken, error)
	GetUserByID(ctx context.Context, id pgtype.UUID) (db.CoreUser, error)
	CreateRefreshToken(ctx context.Context, params db.CreateRefreshTokenParams) (db.CoreRefreshToken, error)
}
