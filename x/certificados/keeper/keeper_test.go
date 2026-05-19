package keeper

import (
	"bytes"
	"context"
	"testing"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/buynnex-corp/byx/x/certificados/types"
	lojastypes "github.com/buynnex-corp/byx/x/lojas/types"
)

type fixture struct {
	ctx          context.Context
	keeper       Keeper
	addressCodec address.Codec
	bank         *mockBankKeeper
	lojas        *mockLojasKeeper
	issuerAddr   sdk.AccAddress
	otherAddr    sdk.AccAddress
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_certificados")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	bank := newMockBankKeeper()
	lojasKeeper := newMockLojasKeeper()

	k := NewKeeper(
		storeService,
		cdc,
		addressCodec,
		authority,
		bank,
		lojasKeeper,
	)

	require.NoError(t, k.ParamsStore.Set(ctx, types.DefaultParams()))

	issuerAddr := sdk.AccAddress(bytes.Repeat([]byte{0x01}, 20))
	issuerStr, err := addressCodec.BytesToString(issuerAddr)
	require.NoError(t, err)
	lojasKeeper.merchants[1] = lojastypes.Merchant{Id: 1, Creator: issuerStr}

	bank.setBalance(issuerAddr, sdk.NewCoins(sdk.NewInt64Coin(lojastypes.DenomBYX, 1_000_000)))

	otherAddr := sdk.AccAddress(bytes.Repeat([]byte{0x02}, 20))

	return &fixture{
		ctx:          ctx,
		keeper:       k,
		addressCodec: addressCodec,
		bank:         bank,
		lojas:        lojasKeeper,
		issuerAddr:   issuerAddr,
		otherAddr:    otherAddr,
	}
}

type mockBankKeeper struct {
	balances       map[string]int64
	moduleBalances map[string]int64
}

func newMockBankKeeper() *mockBankKeeper {
	return &mockBankKeeper{
		balances:       make(map[string]int64),
		moduleBalances: make(map[string]int64),
	}
}

func (m *mockBankKeeper) setBalance(addr sdk.AccAddress, coins sdk.Coins) {
	m.balances[addr.String()] = coins.AmountOf(lojastypes.DenomBYX).Int64()
}

func (m *mockBankKeeper) SendCoinsFromAccountToModule(_ context.Context, from sdk.AccAddress, module string, amt sdk.Coins) error {
	value := amt.AmountOf(lojastypes.DenomBYX).Int64()
	if m.balances[from.String()] < value {
		return sdkerrors.ErrInsufficientFunds.Wrap("insufficient funds")
	}
	m.balances[from.String()] -= value
	m.moduleBalances[module] += value
	return nil
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, module string, to sdk.AccAddress, amt sdk.Coins) error {
	value := amt.AmountOf(lojastypes.DenomBYX).Int64()
	if m.moduleBalances[module] < value {
		return sdkerrors.ErrInsufficientFunds.Wrap("module insufficient funds")
	}
	m.moduleBalances[module] -= value
	m.balances[to.String()] += value
	return nil
}

type mockLojasKeeper struct {
	merchants map[uint64]lojastypes.Merchant
}

func newMockLojasKeeper() *mockLojasKeeper {
	return &mockLojasKeeper{merchants: make(map[uint64]lojastypes.Merchant)}
}

func (m *mockLojasKeeper) GetMerchant(_ context.Context, id uint64) (lojastypes.Merchant, error) {
	merchant, ok := m.merchants[id]
	if !ok {
		return lojastypes.Merchant{}, collections.ErrNotFound
	}
	return merchant, nil
}
