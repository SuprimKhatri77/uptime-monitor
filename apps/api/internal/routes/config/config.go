package config

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/config"
	db "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
	"github.com/suprimkhatri77/uptime-monitor/api/internal/packages/cloudinary"
)

type Config struct {
	Config    *config.Config
	Queries   *db.Queries
	CldClient *cloudinary.Client
	PgxPool   *pgxpool.Pool
}
