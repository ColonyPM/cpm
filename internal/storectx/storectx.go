// internal/storectx/storectx.go
package storectx

import (
	"context"
	"database/sql"

	store "github.com/ColonyPM/cpm/internal/db"
	colonies "github.com/colonyos/colonies/pkg/client"
)

type ctxKey string

const (
	dbKey       ctxKey = "db"
	qKey        ctxKey = "queries"
	coloniesKey ctxKey = "coloniesClient"
)

// WithStore attaches db, queries, and colonies client to a context and returns the new context.
func WithStore(ctx context.Context, db *sql.DB, q *store.Queries, cc *colonies.ColoniesClient) context.Context {
	ctx = context.WithValue(ctx, dbKey, db)
	ctx = context.WithValue(ctx, qKey, q)
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
