package keeper_test

import (
	"testing"
	"time"

	"github.com/buynnex-corp/byx/x/payments/keeper"
	"github.com/buynnex-corp/byx/x/payments/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
)

func TestQueryPaymentRequestMarksExpired(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)
	q := keeper.NewQueryServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:          merchantStr,
		LojaId:           1,
		AmountMicrobyx:   1000,
		ExpiresInSeconds: 60,
	})
	require.NoError(t, err)

	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	f.ctx = sdkCtx.WithBlockTime(sdkCtx.BlockTime().Add(120 * time.Second))

	resp, err := q.PaymentRequest(f.ctx, &types.QueryGetPaymentRequestRequest{Id: createResp.Id})
	require.NoError(t, err)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_EXPIRED, resp.PaymentRequest.Status)

	pr, found := f.keeper.GetPaymentRequest(sdk.UnwrapSDKContext(f.ctx), createResp.Id)
	require.True(t, found)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_EXPIRED, pr.Status)

	listResp, err := q.PaymentRequestsByLoja(f.ctx, &types.QueryPaymentRequestsByLojaRequest{
		LojaId:     1,
		Pagination: &sdkquery.PageRequest{Limit: 5},
	})
	require.NoError(t, err)
	require.NotEmpty(t, listResp.PaymentRequests)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_EXPIRED, listResp.PaymentRequests[0].Status)
}

func TestQueryPaymentRequestsByLojaReturnsCreated(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)
	q := keeper.NewQueryServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        merchantStr,
		LojaId:         1,
		AmountMicrobyx: 1234,
	})
	require.NoError(t, err)

	resp, err := q.PaymentRequestsByLoja(f.ctx, &types.QueryPaymentRequestsByLojaRequest{
		LojaId:     1,
		Pagination: &sdkquery.PageRequest{Limit: 5, Reverse: true},
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.PaymentRequests)
	found := false
	for _, pr := range resp.PaymentRequests {
		if pr.Id == createResp.Id {
			found = true
			require.Equal(t, uint64(1234), pr.AmountMicrobyx)
			require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_PENDING, pr.Status)
		}
	}
	require.True(t, found, "created payment request not found in by_loja query")
}
