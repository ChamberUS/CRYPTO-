package keeper

import (
	"context"
	"errors"
	"strings"

	"cosmossdk.io/collections"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

// Certificate retorna um certificado pelo ID.
func (k Keeper) Certificate(ctx context.Context, req *types.QueryGetCertificateRequest) (*types.QueryGetCertificateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	cert, err := k.getCertificate(ctx, req.Id)
	if err != nil {
		if errors.Is(err, types.ErrCertificateNotFound) {
			return nil, status.Error(codes.NotFound, "certificate not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetCertificateResponse{Certificate: &cert}, nil
}

// CertificatesByOwner retorna certificados paginados por owner.
func (k Keeper) CertificatesByOwner(ctx context.Context, req *types.QueryCertificatesByOwnerRequest) (*types.QueryCertificatesByOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Owner == "" {
		return nil, status.Error(codes.InvalidArgument, "owner is required")
	}

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &sdkquery.PageRequest{Limit: 10, Reverse: true}
	} else if pageReq.Limit == 0 {
		pageReq.Limit = 10
	}

	certs, page, err := sdkquery.CollectionPaginate(
		ctx,
		k.ByOwner,
		pageReq,
		func(key collections.Pair[string, uint64], _ collections.NoValue) (*types.Certificate, error) {
			c, err := k.Certificates.Get(ctx, key.K2())
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
		sdkquery.WithCollectionPaginationPairPrefix[string, uint64](req.Owner),
	)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	filtered := make([]*types.Certificate, 0, len(certs))
	for _, c := range certs {
		if c != nil {
			filtered = append(filtered, c)
		}
	}

	return &types.QueryCertificatesByOwnerResponse{
		Certificates: filtered,
		Pagination:   page,
	}, nil
}

// CertificatesByMerchant retorna certificados por merchant_id.
func (k Keeper) CertificatesByMerchant(ctx context.Context, req *types.QueryCertificatesByMerchantRequest) (*types.QueryCertificatesByMerchantResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.MerchantId == 0 {
		return nil, status.Error(codes.InvalidArgument, "merchant_id is required")
	}

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &sdkquery.PageRequest{Limit: 10, Reverse: true}
	} else if pageReq.Limit == 0 {
		pageReq.Limit = 10
	}

	certs, page, err := sdkquery.CollectionPaginate(
		ctx,
		k.ByMerchant,
		pageReq,
		func(key collections.Pair[uint64, uint64], _ collections.NoValue) (*types.Certificate, error) {
			c, err := k.Certificates.Get(ctx, key.K2())
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
		sdkquery.WithCollectionPaginationPairPrefix[uint64, uint64](req.MerchantId),
	)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	filtered := make([]*types.Certificate, 0, len(certs))
	for _, c := range certs {
		if c != nil {
			filtered = append(filtered, c)
		}
	}

	return &types.QueryCertificatesByMerchantResponse{
		Certificates: filtered,
		Pagination:   page,
	}, nil
}

// CertificatesBySerial retorna certificados por serial_hash.
func (k Keeper) CertificatesBySerial(ctx context.Context, req *types.QueryCertificatesBySerialRequest) (*types.QueryCertificatesBySerialResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.SerialHash == "" {
		return nil, status.Error(codes.InvalidArgument, "serial_hash is required")
	}

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &sdkquery.PageRequest{Limit: 10, Reverse: true}
	} else if pageReq.Limit == 0 {
		pageReq.Limit = 10
	}

	serial := strings.ToLower(req.SerialHash)
	certs, page, err := sdkquery.CollectionPaginate(
		ctx,
		k.BySerialHash,
		pageReq,
		func(key collections.Pair[string, uint64], _ collections.NoValue) (*types.Certificate, error) {
			c, err := k.Certificates.Get(ctx, key.K2())
			if err != nil {
				return nil, err
			}
			return &c, nil
		},
		sdkquery.WithCollectionPaginationPairPrefix[string, uint64](serial),
	)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	filtered := make([]*types.Certificate, 0, len(certs))
	for _, c := range certs {
		if c != nil {
			filtered = append(filtered, c)
		}
	}

	return &types.QueryCertificatesBySerialResponse{
		Certificates: filtered,
		Pagination:   page,
	}, nil
}

// Params retorna os parâmetros do módulo.
func (k Keeper) Params(ctx context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := k.ensureParams(ctx)
	return &types.QueryParamsResponse{Params: &params}, nil
}
