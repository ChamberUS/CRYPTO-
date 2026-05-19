package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/buynnex-corp/byx/x/payments/types"

	"cosmossdk.io/collections"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

func (q queryServer) PaymentRequest(ctx context.Context, req *types.QueryGetPaymentRequestRequest) (*types.QueryGetPaymentRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pr, found := q.k.GetPaymentRequest(sdkCtx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "payment request not found")
	}

	if err := q.k.ensureCurrentStatus(sdkCtx, &pr); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGetPaymentRequestResponse{PaymentRequest: &pr}, nil
}

func (q queryServer) PaymentRequestsByLoja(ctx context.Context, req *types.QueryPaymentRequestsByLojaRequest) (*types.QueryPaymentRequestsByLojaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.LojaId == 0 {
		return nil, status.Error(codes.InvalidArgument, "loja_id is required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pageReq := req.Pagination
	if pageReq == nil {
		pageReq = &sdkquery.PageRequest{Limit: 10, Reverse: true}
	} else if pageReq.Limit == 0 {
		pageReq.Limit = 10
	}

	baseStore := runtime.KVStoreAdapter(q.k.storeService.OpenKVStore(sdkCtx))
	pfxStore := prefix.NewStore(baseStore, types.PaymentRequestByLojaPrefix)
	lojaStore := prefix.NewStore(pfxStore, sdk.Uint64ToBigEndian(req.LojaId))

	var items []*types.PaymentRequest
	page, err := sdkquery.Paginate(lojaStore, pageReq, func(key []byte, _ []byte) error {
		id := sdk.BigEndianToUint64(key)
		pr, found := q.k.GetPaymentRequest(sdkCtx, id)
		if !found {
			return collections.ErrNotFound
		}
		if err := q.k.ensureCurrentStatus(sdkCtx, &pr); err != nil {
			return err
		}
		items = append(items, &pr)
		return nil
	})
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPaymentRequestsByLojaResponse{
		PaymentRequests: items,
		Pagination:      page,
	}, nil
}

func (q queryServer) PaymentsQRCode(ctx context.Context, req *types.QueryPaymentsQRCodeRequest) (*types.QueryPaymentsQRCodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.RequestId == 0 {
		return nil, status.Error(codes.InvalidArgument, "request_id is required")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pr, found := q.k.GetPaymentRequest(sdkCtx, req.RequestId)
	if !found {
		return nil, status.Error(codes.NotFound, "payment request not found")
	}
	if err := q.k.ensureCurrentStatus(sdkCtx, &pr); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	merchant, err := q.k.lojasKeeper.GetMerchant(ctx, pr.LojaId)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "merchant not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	type qrPayload struct {
		T   string `json:"t"`
		RID string `json:"rid"`
		Amt string `json:"amt"`
		MID uint64 `json:"mid"`
		To  string `json:"to"`
	}

	payloadBz, err := json.Marshal(qrPayload{
		T:   "byxpay",
		RID: strconv.FormatUint(pr.Id, 10),
		Amt: strconv.FormatUint(pr.AmountMicrobyx, 10),
		MID: pr.LojaId,
		To:  merchant.Creator,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryPaymentsQRCodeResponse{
		RequestId:              strconv.FormatUint(pr.Id, 10),
		AmountMicrobyx:         pr.AmountMicrobyx,
		MerchantId:             pr.LojaId,
		MerchantOwner:          merchant.Creator,
		SuggestedCertificateId: req.SuggestedCertificateId,
		QrPayload:              string(payloadBz),
	}, nil
}
