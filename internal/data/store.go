package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/YY404NF/pq-backend/internal/model"
	"github.com/YY404NF/pq-backend/internal/payload"
)

type Store struct {
	db *sql.DB
}

func Open(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS catalog_items (
			record_id INTEGER PRIMARY KEY,
			item_name TEXT NOT NULL,
			category TEXT NOT NULL,
			price_cents INTEGER NOT NULL,
			stock_status TEXT NOT NULL,
			merchant TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			display_order INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS catalog_payload_blocks (
			record_id INTEGER NOT NULL,
			block_index INTEGER NOT NULL,
			block_value_u64 INTEGER NOT NULL,
			PRIMARY KEY (record_id, block_index)
		);`,
		`CREATE TABLE IF NOT EXISTS metadata (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			dataset_version TEXT NOT NULL,
			record_count INTEGER NOT NULL,
			block_count INTEGER NOT NULL,
			domain_size INTEGER NOT NULL,
			generated_at TEXT NOT NULL
		);`,
	}
	for _, stmt := range statements {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ReplaceSeedData(ctx context.Context, items []model.CatalogItem, generatedAt string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM catalog_payload_blocks`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM catalog_items`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM metadata`); err != nil {
		return err
	}

	hash := sha256.New()
	insertItem := `INSERT INTO catalog_items
		(record_id, item_name, category, price_cents, stock_status, merchant, updated_at, display_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	insertBlock := `INSERT INTO catalog_payload_blocks (record_id, block_index, block_value_u64) VALUES (?, ?, ?)`

	for _, item := range items {
		if _, err := tx.ExecContext(
			ctx,
			insertItem,
			item.RecordID,
			item.ItemName,
			item.Category,
			item.PriceCents,
			item.StockStatus,
			item.Merchant,
			item.UpdatedAt,
			item.DisplayOrder,
		); err != nil {
			return err
		}

		blocks := payload.EncodeCatalogItem(item)
		for blockIndex, block := range blocks {
			hash.Write([]byte(fmt.Sprintf("%d:%d:%d;", item.RecordID, blockIndex, block)))
			if _, err := tx.ExecContext(ctx, insertBlock, item.RecordID, blockIndex, int64(block)); err != nil {
				return err
			}
		}
	}

	domainSize := nextPowerOfTwo(uint64(len(items)))
	version := hex.EncodeToString(hash.Sum(nil))[:16]
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO metadata (id, dataset_version, record_count, block_count, domain_size, generated_at)
		 VALUES (1, ?, ?, ?, ?, ?)`,
		version,
		len(items),
		payload.BlockCount,
		domainSize,
		generatedAt,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) Version(ctx context.Context) (model.CatalogVersion, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT dataset_version, record_count, block_count, domain_size FROM metadata WHERE id = 1`,
	)
	var version model.CatalogVersion
	err := row.Scan(&version.DatasetVersion, &version.RecordCount, &version.BlockCount, &version.DomainSize)
	return version, err
}

func (s *Store) CatalogItems(ctx context.Context) ([]model.CatalogItem, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT record_id, item_name, category, price_cents, merchant, updated_at, display_order
		 FROM catalog_items ORDER BY display_order`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.CatalogItem, 0, 32)
	for rows.Next() {
		var item model.CatalogItem
		if err := rows.Scan(
			&item.RecordID,
			&item.ItemName,
			&item.Category,
			&item.PriceCents,
			&item.Merchant,
			&item.UpdatedAt,
			&item.DisplayOrder,
		); err != nil {
			return nil, err
		}
		item.PriceText = payload.FormatPrice(item.PriceCents)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) PayloadBlocks(ctx context.Context, recordCount, blockCount int) ([]uint64, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT record_id, block_index, block_value_u64
		 FROM catalog_payload_blocks
		 ORDER BY record_id, block_index`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make([]uint64, recordCount*blockCount)
	for rows.Next() {
		var recordID int
		var blockIndex int
		var value int64
		if err := rows.Scan(&recordID, &blockIndex, &value); err != nil {
			return nil, err
		}
		values[recordID*blockCount+blockIndex] = uint64(value)
	}
	return values, rows.Err()
}

func nextPowerOfTwo(value uint64) uint64 {
	if value <= 1 {
		return 1
	}
	result := uint64(1)
	for result < value {
		result <<= 1
	}
	return result
}
