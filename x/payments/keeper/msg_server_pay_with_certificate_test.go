package keeper_test

import (
	"strconv"
	"testing"

	certificadostypes "github.com/buynnex-corp/byx/x/certificados/types"
	"github.com/buynnex-corp/byx/x/payments/keeper"
	"github.com/buynnex-corp/byx/x/payments/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestPayWithCertificate(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:    merchantStr,
		LojaId:     1,
		AmountUbyx: 2_000,
		Memo:       "cert",
	})
	require.NoError(t, err)

	f.certificados.certs[1] = certificadostypes.Certificate{
		Id:         1,
		MerchantId: 1,
		Issuer:     merchantStr,
		Owner:      payerStr,
		Revoked:    false,
	}

	_, err = ms.PayWithCertificate(f.ctx, &types.MsgPayWithCertificate{
		RequestId:     strconv.FormatUint(createResp.Id, 10),
		CertificateId: 1,
		Payer:         payerStr,
	})
	require.NoError(t, err)

	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	pr, found := f.keeper.GetPaymentRequest(sdkCtx, createResp.Id)
	require.True(t, found)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_PAID, pr.Status)
	require.Equal(t, payerStr, pr.Payer)
	require.NotZero(t, pr.PaidAtUnix)

	updated := f.certificados.certs[1]
	require.Equal(t, merchantStr, updated.Owner)

	events := sdkCtx.EventManager().Events()
	foundPaid := false
	foundTransfer := false
	foundPaidWithCert := false
	for _, ev := range events {
		switch ev.Type {
		case "byx_payment_request_paid":
			foundPaid = true
		case "certificados_transfer":
			foundTransfer = true
		case "payments_paid_with_certificate":
			foundPaidWithCert = true
		}
	}
	require.True(t, foundPaid, "missing byx_payment_request_paid")
	require.True(t, foundTransfer, "missing certificados_transfer")
	require.True(t, foundPaidWithCert, "missing payments_paid_with_certificate")
}

func TestPayWithCertificateFailsIfNotOwner(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:    merchantStr,
		LojaId:     1,
		AmountUbyx: 100,
	})
	require.NoError(t, err)

	f.certificados.certs[1] = certificadostypes.Certificate{
		Id:         1,
		MerchantId: 1,
		Issuer:     merchantStr,
		Owner:      merchantStr, // not the payer
		Revoked:    false,
	}

	_, err = ms.PayWithCertificate(f.ctx, &types.MsgPayWithCertificate{
		RequestId:     strconv.FormatUint(createResp.Id, 10),
		CertificateId: 1,
		Payer:         payerStr,
	})
	require.ErrorIs(t, err, certificadostypes.ErrOwnerMismatch)

	pr, found := f.keeper.GetPaymentRequest(sdk.UnwrapSDKContext(f.ctx), createResp.Id)
	require.True(t, found)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_PENDING, pr.Status)
	require.Equal(t, merchantStr, f.certificados.certs[1].Owner)
}

func TestPayWithCertificateFailsIfRevoked(t *testing.T) {
	f := initFixture(t)
	ms := keeper.NewMsgServerImpl(f.keeper)

	merchantStr, _ := f.addressCodec.BytesToString(f.merchantAddr)
	payerStr, _ := f.addressCodec.BytesToString(f.payerAddr)

	createResp, err := ms.CreatePaymentRequest(f.ctx, &types.MsgCreatePaymentRequest{
		Creator:    merchantStr,
		LojaId:     1,
		AmountUbyx: 100,
	})
	require.NoError(t, err)

	f.certificados.certs[1] = certificadostypes.Certificate{
		Id:         1,
		MerchantId: 1,
		Issuer:     merchantStr,
		Owner:      payerStr,
		Revoked:    true,
	}

	_, err = ms.PayWithCertificate(f.ctx, &types.MsgPayWithCertificate{
		RequestId:     strconv.FormatUint(createResp.Id, 10),
		CertificateId: 1,
		Payer:         payerStr,
	})
	require.ErrorIs(t, err, certificadostypes.ErrCertificateRevoked)

	pr, found := f.keeper.GetPaymentRequest(sdk.UnwrapSDKContext(f.ctx), createResp.Id)
	require.True(t, found)
	require.Equal(t, types.PaymentStatus_PAYMENT_STATUS_PENDING, pr.Status)
	require.Equal(t, payerStr, f.certificados.certs[1].Owner)
}
