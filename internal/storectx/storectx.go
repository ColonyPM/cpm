// internal/storectx/storectx.go
package storectx

import (
	"context"
	"database/sql"

	"github.com/ColonyPM/cpm/internal/config"
	store "github.com/ColonyPM/cpm/internal/db"
	colonies "github.com/colonyos/colonies/pkg/client"
)

type ctxKey string

const (
	dbKey       ctxKey = "db"
	qKey        ctxKey = "queries"
	coloniesKey ctxKey = "coloniesClient"
	configKey   ctxKey = "config"
)

// WithStore attaches db, queries, colonies client, and config to a context and returns the new context.
func WithStore(ctx context.Context, db *sql.DB, q *store.Queries, cc *colonies.ColoniesClient, cfg *config.Config) context.Context {
	ctx = context.WithValue(ctx, dbKey, db)
	ctx = context.WithValue(ctx, qKey, q)
	ctx = context.WithValue(ctx, configKey, cfg)
	if cc != nil {
		ctx = context.WithValue(ctx, coloniesKey, cc)
	}
	return ctx
}

func GetDb(ctx context.Context) (*sql.DB, *store.Queries) {
	db, _ := ctx.Value(dbKey).(*sql.DB)
	q, _ := ctx.Value(qKey).(*store.Queries)
	if db == nil || q == nil {
		panic("storectx: DB/Queries not initialized on context")
	}
	return db, q
}

func GetColoniesClient(ctx context.Context) *colonies.ColoniesClient {
	cc, _ := ctx.Value(coloniesKey).(*colonies.ColoniesClient)
	if cc == nil {
		panic("storectx: Colonies client not initialized on context")
	}
	return cc
}

func IsInitialized(ctx context.Context) bool {
	_, ok := ctx.Value(dbKey).(*sql.DB)
	return ok
}

func GetConfig(ctx context.Context) *config.Config {
	cfg, _ := ctx.Value(configKey).(*config.Config)
	if cfg == nil {
		panic("storectx: Config not initialized on context")
	}
	return cfg
}
