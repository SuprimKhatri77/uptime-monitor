package app

import (
	"github.com/suprimkhatri77/uptime-monitor/api/internal/cron"
	dbgen "github.com/suprimkhatri77/uptime-monitor/api/internal/database/generated"
)

func initCron(queries *dbgen.Queries) {
	cron.CronExample(queries)
}
