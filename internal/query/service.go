package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/YY404NF/pq-backend/internal/config"
	"github.com/YY404NF/pq-backend/internal/data"
	"github.com/YY404NF/pq-backend/internal/dpfbridge"
	"github.com/YY404NF/pq-backend/internal/model"
)

type EvalRequest struct {
	DatasetVersion string              `json:"datasetVersion"`
	QueryID        string              `json:"queryId"`
	DomainSize     uint64              `json:"domainSize"`
	KeyShare       EvalRequestKeyShare `json:"keyShare"`
}

type EvalRequestKeyShare struct {
	SeedHex         string                      `json:"seedHex"`
	CorrectionWords []EvalRequestCorrectionWord `json:"correctionWords"`
}

type EvalRequestCorrectionWord struct {
	SHex string `json:"sHex"`
	Tr   bool   `json:"tr"`
}

type EvalResponse struct {
	Server               string            `json:"server"`
	Party                int               `json:"party"`
	DatasetVersion       string            `json:"datasetVersion"`
	QueryID              string            `json:"queryId"`
	ResultShareBlocksHex []string          `json:"resultShareBlocksHex"`
	ElapsedMs            int64             `json:"elapsedMs"`
	Trace                EvalResponseTrace `json:"trace"`
}

type EvalResponseTrace struct {
	WorkerCount int `json:"workerCount"`
	RecordCount int `json:"recordCount"`
	BlockCount  int `json:"blockCount"`
}

type Service struct {
	cfg   config.Config
	store *data.Store
}

func NewService(cfg config.Config, store *data.Store) *Service {
	return &Service{cfg: cfg, store: store}
}

func (s *Service) Version(ctx context.Context) (model.CatalogVersion, error) {
	return s.store.Version(ctx)
}

func (s *Service) CatalogItems(ctx context.Context) ([]model.CatalogItem, error) {
	return s.store.CatalogItems(ctx)
}

func (s *Service) PayloadBlocks(ctx context.Context, recordCount, blockCount int) ([]uint64, error) {
	return s.store.PayloadBlocks(ctx, recordCount, blockCount)
}

func (s *Service) BuildKeyShare(req EvalRequest) (dpfbridge.KeyShare, error) {
	version, err := s.store.Version(context.Background())
	if err != nil {
		return dpfbridge.KeyShare{}, err
	}
	if req.DatasetVersion != version.DatasetVersion {
		return dpfbridge.KeyShare{}, ErrVersionMismatch
	}
	if req.DomainSize != version.DomainSize {
		return dpfbridge.KeyShare{}, fmt.Errorf("domain size mismatch")
	}
	seed, err := dpfbridge.DecodeBlock128Hex(req.KeyShare.SeedHex)
	if err != nil {
		return dpfbridge.KeyShare{}, fmt.Errorf("invalid seed hex: %w", err)
	}
	inBits := resolveInBits(req.DomainSize)
	if len(req.KeyShare.CorrectionWords) != inBits+1 {
		return dpfbridge.KeyShare{}, fmt.Errorf("invalid correction word count")
	}
	cws := make([]dpfbridge.CorrectionWord, 0, len(req.KeyShare.CorrectionWords))
	for _, cw := range req.KeyShare.CorrectionWords {
		block, err := dpfbridge.DecodeBlock128Hex(cw.SHex)
		if err != nil {
			return dpfbridge.KeyShare{}, fmt.Errorf("invalid correction word hex: %w", err)
		}
		cws = append(cws, dpfbridge.CorrectionWord{S: block, Tr: cw.Tr})
	}
	return dpfbridge.KeyShare{
		Seed:            seed,
		CorrectionWords: cws,
		InBits:          uint32(inBits),
		DomainSize:      req.DomainSize,
	}, nil
}

func resolveInBits(domainSize uint64) int {
	bits := 0
	for (uint64(1) << bits) < domainSize {
		bits++
	}
	return bits
}

var ErrVersionMismatch = errors.New("dataset version mismatch")
