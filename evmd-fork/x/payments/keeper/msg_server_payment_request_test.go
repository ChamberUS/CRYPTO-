package keeper_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/buynnex/iaos-evmd/x/payments/keeper"
	"github.com/buynnex/iaos-evmd/x/payments/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestCreatePaymentRequest(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	creator, err := f.addressCodec.BytesToString(f.merchantAddr)
	require.NoError(t, err)

	resp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:          creator,
		LojaId:           1,
		AmountMicrobyx:   1500,
		Memo:             "pedido #1",
		ExpiresInSeconds: 120,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Id)

	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	pr, found := f.keeper.GetPaymentRequest(sdkCtx, resp.Id)
	require.True(t, found)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_PENDING, pr.Status)
	require.Equal(t, uint64(1), pr.LojaId)
	require.Equal(t, uint64(1500), pr.AmountMicrobyx)
	require.NotZero(t, pr.ExpiresAtUnix)

	events := sdkCtx.EventManager().Events()
	foundEvent := false
	for _, ev := range events {
		if ev.Type != "byx_payment_request_created" {
			continue
		}
		foundEvent = true
		attrs := make(map[string]string)
		for _, a := range ev.Attributes {
			attrs[a.Key] = string(a.Value)
		}
		require.Equal(t, "1", attrs["request_id"])
		require.Equal(t, "1", attrs["loja_id"])
		require.Equal(t, "1500", attrs["amount_microbyx"])
		require.NotEmpty(t, attrs["expires_at_unix"])
		require.NotEmpty(t, attrs["fingerprint_hash"])
	}
	require.True(t, foundEvent, "created event not emitted")
}

func TestCreatePaymentRequestDedupReusesPending(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	creator, _ := f.addressCodec.BytesToString(f.merchantAddr)

	first, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        creator,
		LojaId:         1,
		AmountMicrobyx: 500,
		Memo:           "A",
	})
	require.NoError(t, err)

	second, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        creator,
		LojaId:         1,
		AmountMicrobyx: 500,
		Memo:           "A", // same fingerprint
	})
	require.NoError(t, err)
	require.Equal(t, first.Id, second.Id, "should reuse pending request")

	_, found := f.keeper.GetDedupeRequestID(sdk.UnwrapSDKContext(f.ctx), 1, 500, "A")
	require.True(t, found)
}

func TestPayPaymentRequest(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        merchantStr,
		LojaId:         1,
		AmountMicrobyx: 2_000,
	})
	require.NoError(t, err)

	_, err = ms.PayPaymentRequest(f.ctx, &types.MsgPayPaymentRequest{
		Creator:   payerStr,
		RequestId: createResp.Id,
	})
	require.NoError(t, err)

	pr, found := f.keeper.GetPaymentRequest(sdk.UnwrapSDKContext(f.ctx), createResp.Id)
	require.True(t, found)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_PAID, pr.Status)
	require.Equal(t, payerStr, pr.Payer)
	require.NotZero(t, pr.PaidAtUnix)

	require.Equal(t, int64(998000), f.bank.balances[payerStr])
	require.Equal(t, int64(2000), f.bank.balances[merchantStr])

	// dedupe must be cleared after pay
	_, found = f.keeper.GetDedupeRequestID(sdk.UnwrapSDKContext(f.ctx), 1, 2_000, "")
	require.False(t, found, "dedupe index should be cleared on pay")
}

func TestPayPaymentRequestExpired(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:          merchantStr,
		LojaId:           1,
		AmountMicrobyx:   500,
		ExpiresInSeconds: 60,
	})
	require.NoError(t, err)

	// advance time past expiration
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	f.ctx = sdkCtx.WithBlockTime(sdkCtx.BlockTime().Add(120 * time.Second))

	_, err = ms.PayPaymentRequest(f.ctx, &types.MsgPayPaymentRequest{
		Creator:   payerStr,
		RequestId: createResp.Id,
	})
	require.ErrorIs(t, err, types.ErrPaymentRequestExpired)

	pr, found := f.keeper.GetPaymentRequest(sdk.UnwrapSDKContext(f.ctx), createResp.Id)
	require.True(t, found)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_EXPIRED, pr.Status)

	// creating again after expiration should yield a new ID
	newResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:          merchantStr,
		LojaId:           1,
		AmountMicrobyx:   500,
		Memo:             "",
		ExpiresInSeconds: 60,
	})
	require.NoError(t, err)
	require.NotEqual(t, createResp.Id, newResp.Id)
}

func TestCreatePaymentRequestAfterPaidCreatesNewID(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	first, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        merchantStr,
		LojaId:         1,
		AmountMicrobyx: 123,
		Memo:           "same",
	})
	require.NoError(t, err)

	_, err = ms.PayPaymentRequest(f.ctx, &types.MsgPayPaymentRequest{
		Creator:   payerStr,
		RequestId: first.Id,
	})
	require.NoError(t, err)

	second, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        merchantStr,
		LojaId:         1,
		AmountMicrobyx: 123,
		Memo:           "same",
	})
	require.NoError(t, err)
	require.NotEqual(t, first.Id, second.Id)
}

func TestPayPaymentRequestTwice(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        merchantStr,
		LojaId:         1,
		AmountMicrobyx: 700,
	})
	require.NoError(t, err)

	_, err = ms.PayPaymentRequest(f.ctx, &types.MsgPayPaymentRequest{
		Creator:   payerStr,
		RequestId: createResp.Id,
	})
	require.NoError(t, err)

	_, err = ms.PayPaymentRequest(f.ctx, &types.MsgPayPaymentRequest{
		Creator:   payerStr,
		RequestId: createResp.Id,
	})
	require.ErrorIs(t, err, types.ErrPaymentRequestNotPending)
}

func TestPayPaymentRequestEmitsEnrichedEvent(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:        merchantStr,
		LojaId:         1,
		AmountMicrobyx: 700,
		Memo:           "evento",
	})
	require.NoError(t, err)

	_, err = ms.PayPaymentRequest(f.ctx, &types.MsgPayPaymentRequest{
		Creator:   payerStr,
		RequestId: createResp.Id,
	})
	require.NoError(t, err)

	events := sdk.UnwrapSDKContext(f.ctx).EventManager().Events()
	found := false

	for _, ev := range events {
		if ev.Type != "byx_payment_request_paid" {
			continue
		}
		found = true
		attrs := make(map[string]string)
		for _, a := range ev.Attributes {
			attrs[a.Key] = string(a.Value)
		}

		require.Equal(t, strconv.FormatUint(createResp.Id, 10), attrs["request_id"])
		require.Equal(t, strconv.FormatUint(1, 10), attrs["loja_id"])
		require.Equal(t, payerStr, attrs["payer"])
		require.Equal(t, strconv.FormatUint(700, 10), attrs["amount_microbyx"])
		require.NotEmpty(t, attrs["paid_at_unix"])
	}

	require.True(t, found, "payment event not found")
}
