package httpapi

import (
	"context"
	"errors"
	"time"

	"github.com/YY404NF/pq-backend/internal/config"
	"github.com/YY404NF/pq-backend/internal/dpfbridge"
	"github.com/YY404NF/pq-backend/internal/query"
)

var ErrBadRequest = errors.New("bad request")

func Eval(ctx context.Context, cfg config.Config, service *query.Service, req query.EvalRequest) (query.EvalResponse, error) {
	version, err := service.Version(ctx)
	if err != nil {
		return query.EvalResponse{}, err
	}
	if req.DatasetVersion != version.DatasetVersion {
		return query.EvalResponse{}, query.ErrVersionMismatch
	}

	key, err := service.BuildKeyShare(req)
	if err != nil {
		return query.EvalResponse{}, errors.Join(ErrBadRequest, err)
	}

	payload, err := service.PayloadBlocks(ctx, version.RecordCount, version.BlockCount)
	if err != nil {
		return query.EvalResponse{}, err
	}

	start := time.Now()
	share, err := dpfbridge.AggregateQueryShare(
		cfg.Party,
		key,
		payload,
		version.RecordCount,
		version.BlockCount,
		cfg.WorkerCount,
	)
	if err != nil {
		return query.EvalResponse{}, err
	}
	hexBlocks := make([]string, len(share))
	for i, block := range share {
		hexBlocks[i] = dpfbridge.EncodeU64Hex(block)
	}

	return query.EvalResponse{
		Server:               cfg.ServerName,
		Party:                cfg.Party,
		DatasetVersion:       version.DatasetVersion,
		QueryID:              req.QueryID,
		ResultShareBlocksHex: hexBlocks,
		ElapsedMs:            time.Since(start).Milliseconds(),
		Trace: query.EvalResponseTrace{
			WorkerCount: cfg.WorkerCount,
			RecordCount: version.RecordCount,
			BlockCount:  version.BlockCount,
		},
	}, nil
}
