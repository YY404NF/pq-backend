package app

import (
	"context"
	"log"
	"time"

	"github.com/YY404NF/pq-backend/internal/config"
	"github.com/YY404NF/pq-backend/internal/data"
	"github.com/YY404NF/pq-backend/internal/httpapi"
	"github.com/YY404NF/pq-backend/internal/query"
	"github.com/YY404NF/pq-backend/internal/sample"
)

func Run(cfg config.Config) error {
	store, err := data.Open(cfg.DBPath)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.EnsureSchema(ctx); err != nil {
		return err
	}
	if err := store.ReplaceSeedData(ctx, sample.CatalogItems(), time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	service := query.NewService(cfg, store)
	router := httpapi.NewRouter(cfg, service)

	log.Printf("%s listening on %s with db=%s party=%d workers=%d", cfg.ServerName, cfg.Address(), cfg.DBPath, cfg.Party, cfg.WorkerCount)
	return router.Run(cfg.Address())
}
