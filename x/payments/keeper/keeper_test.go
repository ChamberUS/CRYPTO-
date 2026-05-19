package keeper_test

import (
	"bytes"
	"context"
	"strconv"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	certificadostypes "github.com/buynnex-corp/byx/x/certificados/types"
	lojas "github.com/buynnex-corp/byx/x/lojas/types"
	"github.com/buynnex-corp/byx/x/payments/keeper"
	module "github.com/buynnex-corp/byx/x/payments/module"
	"github.com/buynnex-corp/byx/x/payments/types"
)

type fixture struct {
	ctx          context.Context
	keeper       keeper.Keeper
	addressCodec address.Codec
	bank         *mockBankKeeper
	lojas        *mockLojasKeeper
	certificados *mockCertificadosKeeper
	merchantAddr sdk.AccAddress
	payerAddr    sdk.AccAddress
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	bank := newMockBankKeeper()
	lojasKeeper := newMockLojasKeeper()
	certificadosKeeper := newMockCertificadosKeeper()

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addressCodec,
		authority,
		bank,
		lojasKeeper,
		certificadosKeeper,
	)

	// Initialize params
	if err := k.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	merchantAddr := sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))
	merchantStr, err := addressCodec.BytesToString(merchantAddr)
	if err != nil {
		t.Fatalf("failed to convert merchant addr: %v", err)
	}
	lojasKeeper.merchants[1] = lojas.Merchant{
		Id:      1,
		Creator: merchantStr,
	}

	payerAddr := sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20))
	bank.setBalance(payerAddr, sdk.NewCoins(sdk.NewInt64Coin(lojas.DenomBYX, 1_000_000)))

	return &fixture{
		ctx:          ctx,
		keeper:       k,
		addressCodec: addressCodec,
		bank:         bank,
		lojas:        lojasKeeper,
		certificados: certificadosKeeper,
		merchantAddr: merchantAddr,
		payerAddr:    payerAddr,
	}
}

type mockBankKeeper struct {
	balances map[string]int64
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{balances: make(map[string]int64)}
}

func (m *mockBankKeeper) setBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.balances[addr.String()] = coins.AmountOf(lojas.DenomBYX).Int64()
}

func (m *mockBankKeeper) SendCoins(_ context.Context, from sdk.AccAddress, to sdk.AccAddress, amt sdk.Coins) error {
	value := amt.AmountOf(lojas.DenomBYX).Int64()
	if m.balances[from.String()] < value {
		return sdkerrors.ErrInsufficientFunds.Wrapf("insufficient funds: %d < %d", m.balances[from.String()], value)
	}
	m.balances[from.String()] -= value
	m.balances[to.String()] += value
	return nil
}

type mockLojasKeeper struct {
	merchants map[uint64]lojas.Merchant
}

func newMockLojasKeeper() *mockLojasKeeper {
	return &mockLojasKeeper{merchants: make(map[uint64]lojas.Merchant)}
}

func (m *mockLojasKeeper) GetMerchant(_ context.Context, id uint64) (lojas.Merchant, error) {
	merchant, ok := m.merchants[id]
	if !ok {
		return lojas.Merchant{}, collections.ErrNotFound
	}
	return merchant, nil
}

type mockCertificadosKeeper struct {
	params certificadostypes.Params
	certs  map[uint64]certificadostypes.Certificate
}

func newMockCertificadosKeeper() *mockCertificadosKeeper {
	return &mockCertificadosKeeper{
		params: certificadostypes.DefaultParams(),
		certs:  make(map[uint64]certificadostypes.Certificate),
	}
}

func (m *mockCertificadosKeeper) GetCertificate(_ context.Context, id uint64) (certificadostypes.Certificate, error) {
	c, ok := m.certs[id]
	if !ok {
		return certificadostypes.Certificate{}, certificadostypes.ErrCertificateNotFound
	}
	return c, nil
}

func (m *mockCertificadosKeeper) GetParams(_ context.Context) (certificadostypes.Params, error) {
	return m.params, nil
}

func (m *mockCertificadosKeeper) TransferCertificate(ctx context.Context, id uint64, from, to string) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if !m.params.Enabled {
		return certificadostypes.ErrModuleDisabled
	}
	if !m.params.AllowTransfer {
		return certificadostypes.ErrTransferNotAllowed
	}

	c, ok := m.certs[id]
	if !ok {
		return certificadostypes.ErrCertificateNotFound
	}
	if c.Revoked {
		return certificadostypes.ErrCertificateRevoked
	}
	if c.Owner != from {
		return certificadostypes.ErrOwnerMismatch
	}
	c.Owner = to
	m.certs[id] = c

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			"certificados_transfer",
			sdk.NewAttribute("id", strconv.FormatUint(id, 10)),
			sdk.NewAttribute("from", from),
			sdk.NewAttribute("to", to),
		),
	)

	return nil
}
