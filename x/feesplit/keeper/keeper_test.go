package keeper_test

import (
	"context"
	"testing"

	"github.com/buynnex-corp/byx/x/feesplit/keeper"
	"github.com/buynnex-corp/byx/x/feesplit/types"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type fixture struct {
	ctx    context.Context
	keeper keeper.Keeper
	bank   *mockBank
	cdc    codec.Codec
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig()
	cdc := encCfg.Codec
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	sb := collections.NewSchemaBuilder(storeService)
	if _, err := sb.Build(); err != nil {
		t.Fatalf("failed to build schema: %v", err)
	}

	bank := newMockBank()
	k := keeper.NewKeeper(
		storeService,
		cdc,
		bank,
	)

	if err := k.SetParams(sdk.UnwrapSDKContext(ctx), types.DefaultParams()); err != nil {
		t.Fatalf("failed to set default params: %v", err)
	}

	return &fixture{
		ctx:    ctx,
		keeper: k,
		bank:   bank,
		cdc:    cdc,
	}
}

func TestBeginBlockerBasicSplit(t *testing.T) {
	f := initFixture(t)

	feeAmt := sdk.NewInt64Coin(types.DefaultDenom, 10_000)
	f.bank.setModuleBalance(authtypes.FeeCollectorName, sdk.NewCoins(feeAmt))

	supplyBefore := f.bank.totalSupply(types.DefaultDenom)

	err := f.keeper.BeginBlocker(sdk.UnwrapSDKContext(f.ctx))
	if err != nil {
		t.Fatalf("begin blocker failed: %v", err)
	}

	treasuryBal := f.bank.balanceForModule(types.TreasuryModuleAccount, types.DefaultDenom)
	if treasuryBal != 3_000 {
		t.Fatalf("treasury balance mismatch: got %d want 3000", treasuryBal)
	}

	feeCollectorBal := f.bank.balanceForModule(authtypes.FeeCollectorName, types.DefaultDenom)
	if feeCollectorBal != 6_000 {
		t.Fatalf("fee collector balance mismatch: got %d want 6000", feeCollectorBal)
	}

	supplyAfter := f.bank.totalSupply(types.DefaultDenom)
	if supplyBefore-supplyAfter != 1_000 {
		t.Fatalf("burn mismatch: supply delta %d want 1000", supplyBefore-supplyAfter)
	}
}

func TestBeginBlockerAllowlistSkipsOtherDenom(t *testing.T) {
	f := initFixture(t)

	customParams := types.DefaultParams()
	customParams.DenomsAllowlist = []string{types.DefaultDenom}
	if err := f.keeper.SetParams(sdk.UnwrapSDKContext(f.ctx), customParams); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	f.bank.setModuleBalance(authtypes.FeeCollectorName, sdk.NewCoins(sdk.NewInt64Coin("other", 5_000)))
	supplyBefore := f.bank.totalSupply("other")

	if err := f.keeper.BeginBlocker(sdk.UnwrapSDKContext(f.ctx)); err != nil {
		t.Fatalf("begin blocker failed: %v", err)
	}

	if got := f.bank.balanceForModule(types.TreasuryModuleAccount, "other"); got != 0 {
		t.Fatalf("treasury should not receive other denom, got %d", got)
	}
	if got := f.bank.balanceForModule(authtypes.FeeCollectorName, "other"); got != 5_000 {
		t.Fatalf("fee collector should stay with other denom, got %d", got)
	}
	if supplyBefore != f.bank.totalSupply("other") {
		t.Fatalf("supply should remain unchanged for other denom")
	}
}

func TestParamsInvalidSum(t *testing.T) {
	p := types.Params{
		Enabled:            true,
		SplitBpsValidators: 5000,
		SplitBpsTreasury:   5000,
		SplitBpsBurn:       1001, // total 11001 invalid
		DenomsAllowlist:    []string{types.DefaultDenom},
	}
	if err := p.Validate(); err == nil {
		t.Fatalf("expected validate to fail when sum != 10000")
	}
}

type mockBank struct {
	balances map[string]sdk.Coins
	supply   map[string]sdkmath.Int
}

func newMockBank() *mockBank {
	return &mockBank{
		balances: make(map[string]sdk.Coins),
		supply:   make(map[string]sdkmath.Int),
	}
}

func (m *mockBank) setModuleBalance(module string, coins sdk.Coins) {
	addr := authtypes.NewModuleAddress(module)
	m.balances[addr.String()] = coins
	m.rebuildSupply()
}

func (m *mockBank) balanceForModule(module, denom string) int64 {
	addr := authtypes.NewModuleAddress(module).String()
	coins := m.balances[addr]
	return coins.AmountOf(denom).Int64()
}

func (m *mockBank) totalSupply(denom string) int64 {
	return m.supply[denom].Int64()
}

func (m *mockBank) rebuildSupply() {
	m.supply = make(map[string]sdkmath.Int)
	for _, coins := range m.balances {
		for _, c := range coins {
			curr, ok := m.supply[c.Denom]
			if !ok {
				curr = sdkmath.NewInt(0)
			}
			m.supply[c.Denom] = curr.Add(c.Amount)
		}
	}
}

func (m *mockBank) GetAllBalances(_ context.Context, addr sdk.AccAddress) sdk.Coins {
	return m.balances[addr.String()]
}

func (m *mockBank) SendCoinsFromModuleToModule(_ context.Context, senderModule string, recipientModule string, amt sdk.Coins) error {
	fromAddr := authtypes.NewModuleAddress(senderModule).String()
	toAddr := authtypes.NewModuleAddress(recipientModule).String()

	fromBal := m.balances[fromAddr]
	if !fromBal.IsAllGTE(amt) {
		return sdkerrors.ErrInsufficientFunds
	}
	m.balances[fromAddr] = fromBal.Sub(amt...)
	m.balances[toAddr] = m.balances[toAddr].Add(amt...)
	m.rebuildSupply()
	return nil
}

func (m *mockBank) BurnCoins(_ context.Context, moduleName string, amt sdk.Coins) error {
	addr := authtypes.NewModuleAddress(moduleName).String()
	bal := m.balances[addr]
	if !bal.IsAllGTE(amt) {
		return sdkerrors.ErrInsufficientFunds
	}
	m.balances[addr] = bal.Sub(amt...)
	for _, c := range amt {
		m.supply[c.Denom] = m.supply[c.Denom].Sub(c.Amount)
	}
	return nil
}
