package keeper

import (
	"testing"

	"github.com/buynnex-corp/byx/x/certificados/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestIssueCertificateWithFee(t *testing.T) {
	f := initFixture(t)
	ms := NewMsgServerImpl(f.keeper)

	issuerStr, _ := f.addressCodec.BytesToString(f.issuerAddr)

	resp, err := ms.IssueCertificate(f.ctx, &types.MsgIssueCertificate{
		Creator:     issuerStr,
		MerchantId:  1,
		Owner:       issuerStr,
		Category:    "NOTEBOOK",
		Brand:       "Lenovo",
		Model:       "X1",
		SerialHash:  "deadbeef",
		Condition:   "A+",
		ImageUri:    "file:///tmp/x1.png",
		ImageSha256: "abc123",
		ImageSeed:   "seed-1",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), resp.Id)

	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	cert, err := f.keeper.Certificates.Get(sdkCtx, resp.Id)
	require.NoError(t, err)
	require.Equal(t, issuerStr, cert.Owner)
	require.Equal(t, uint64(1), cert.MerchantId)

	require.Equal(t, int64(1_000_000)-int64(types.DefaultIssueFeeByx), f.bank.balances[issuerStr])
	require.Equal(t, int64(types.DefaultIssueFeeByx), f.bank.moduleBalances[types.ModuleName])
}

func TestIssueCertificateFailsIfNotMerchantOwner(t *testing.T) {
	f := initFixture(t)
	ms := NewMsgServerImpl(f.keeper)

	issuerStr, _ := f.addressCodec.BytesToString(f.otherAddr)

	_, err := ms.IssueCertificate(f.ctx, &types.MsgIssueCertificate{
		Creator:     issuerStr,
		MerchantId:  1,
		Owner:       issuerStr,
		Category:    "GPU",
		Brand:       "Nvidia",
		Model:       "4090",
		SerialHash:  "hash",
		Condition:   "A",
		ImageUri:    "file:///tmp/gpu.png",
		ImageSha256: "hash2",
		ImageSeed:   "seed",
	})
	require.ErrorIs(t, err, types.ErrNotMerchantOwner)
}

func TestTransferUpdatesOwnerAndIndex(t *testing.T) {
	f := initFixture(t)
	ms := NewMsgServerImpl(f.keeper)

	issuerStr, _ := f.addressCodec.BytesToString(f.issuerAddr)
	otherStr, _ := f.addressCodec.BytesToString(f.otherAddr)

	resp, err := ms.IssueCertificate(f.ctx, &types.MsgIssueCertificate{
		Creator:     issuerStr,
		MerchantId:  1,
		Owner:       issuerStr,
		Category:    "MOUSE",
		Brand:       "Logi",
		Model:       "GPRO",
		SerialHash:  "mousehash",
		Condition:   "A",
		ImageUri:    "file:///tmp/mouse.png",
		ImageSha256: "sha",
		ImageSeed:   "seed",
	})
	require.NoError(t, err)

	_, err = ms.TransferCertificate(f.ctx, &types.MsgTransferCertificate{
		Creator:       issuerStr,
		CertificateId: resp.Id,
		NewOwner:      otherStr,
	})
	require.NoError(t, err)

	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	updated, err := f.keeper.Certificates.Get(sdkCtx, resp.Id)
	require.NoError(t, err)
	require.Equal(t, otherStr, updated.Owner)

	qResp, err := f.keeper.CertificatesByOwner(sdk.WrapSDKContext(sdkCtx), &types.QueryCertificatesByOwnerRequest{
		Owner: otherStr,
	})
	require.NoError(t, err)
	require.Len(t, qResp.Certificates, 1)
}

func TestRevokeBlocksTransfer(t *testing.T) {
	f := initFixture(t)
	ms := NewMsgServerImpl(f.keeper)

	issuerStr, _ := f.addressCodec.BytesToString(f.issuerAddr)
	otherStr, _ := f.addressCodec.BytesToString(f.otherAddr)

	resp, err := ms.IssueCertificate(f.ctx, &types.MsgIssueCertificate{
		Creator:     issuerStr,
		MerchantId:  1,
		Owner:       issuerStr,
		Category:    "GPU",
		Brand:       "XFX",
		Model:       "6800",
		SerialHash:  "gpu",
		Condition:   "B",
		ImageUri:    "file:///tmp/gpu.png",
		ImageSha256: "hash",
		ImageSeed:   "seed",
	})
	require.NoError(t, err)

	_, err = ms.RevokeCertificate(f.ctx, &types.MsgRevokeCertificate{
		Creator:       issuerStr,
		CertificateId: resp.Id,
		Reason:        "fraud",
	})
	require.NoError(t, err)

	_, err = ms.TransferCertificate(f.ctx, &types.MsgTransferCertificate{
		Creator:       issuerStr,
		CertificateId: resp.Id,
		NewOwner:      otherStr,
	})
	require.ErrorIs(t, err, types.ErrCertificateRevoked)
}

func TestAddServiceRecordAddsEntry(t *testing.T) {
	f := initFixture(t)
	ms := NewMsgServerImpl(f.keeper)

	issuerStr, _ := f.addressCodec.BytesToString(f.issuerAddr)

	resp, err := ms.IssueCertificate(f.ctx, &types.MsgIssueCertificate{
		Creator:     issuerStr,
		MerchantId:  1,
		Owner:       issuerStr,
		Category:    "NOTEBOOK",
		Brand:       "Dell",
		Model:       "XPS",
		SerialHash:  "serv-hash",
		Condition:   "A",
		ImageUri:    "file:///tmp/xps.png",
		ImageSha256: "hash",
		ImageSeed:   "seed",
	})
	require.NoError(t, err)

	addResp, err := ms.AddServiceRecord(f.ctx, &types.MsgAddServiceRecord{
		Creator:       issuerStr,
		CertificateId: resp.Id,
		Kind:          "CLEANING",
		Details:       "Dust removal",
	})
	require.NoError(t, err)
	require.Equal(t, uint64(1), addResp.Record.Index)

	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	count, err := f.keeper.ServiceRecordsCount.Get(sdkCtx, resp.Id)
	require.NoError(t, err)
	require.Equal(t, uint64(1), count)
}
