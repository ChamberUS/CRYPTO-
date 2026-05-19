package keeper

import (
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	corestore "cosmossdk.io/core/store"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/buynnex/iaos-evmd/x/feesplit/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	bankKeeper   types.BankKeeper

	Schema collections.Schema
	Params collections.Item[types.Params]
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	bk types.BankKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		bankKeeper:   bk,
		Params:       collections.NewItem(sb, types.ParamsKey, "params", paramsValueCodec{}),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

func (k Keeper) ModuleName() string {
	return types.ModuleName
}

func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.DefaultParams(), nil
		}
		return types.Params{}, err
	}
	return params, nil
}

func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return k.Params.Set(ctx, params)
}

// BeginBlocker applies the fee split logic before distribution consumes the fee collector.
func (k Keeper) BeginBlocker(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if !params.Enabled {
		return nil
	}

	feeCollectorAddr := authtypes.NewModuleAddress(authtypes.FeeCollectorName)

	fees := k.bankKeeper.GetAllBalances(ctx, feeCollectorAddr)
	if fees.IsZero() {
		return nil
	}

	applyAllDenoms := len(params.DenomsAllowlist) == 0
	allowed := make(map[string]struct{}, len(params.DenomsAllowlist))
	for _, denom := range params.DenomsAllowlist {
		allowed[denom] = struct{}{}
	}

	// treasury module account is created at init via module account perms; address is deterministic.

	for _, coin := range fees {
		if !applyAllDenoms {
			if _, ok := allowed[coin.Denom]; !ok {
				continue
			}
		}
		if coin.Amount.IsZero() {
			continue
		}

		total := coin.Amount
		treasuryAmt := floorPortion(total, params.SplitBpsTreasury)
		burnAmt := floorPortion(total, params.SplitBpsBurn)

		if treasuryAmt.IsPositive() {
			sendCoins := sdk.NewCoins(sdk.NewCoin(coin.Denom, treasuryAmt))
			if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, types.TreasuryModuleAccount, sendCoins); err != nil {
				return err
			}
		}

		if burnAmt.IsPositive() {
			burnCoins := sdk.NewCoins(sdk.NewCoin(coin.Denom, burnAmt))
			if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.FeeCollectorName, types.ModuleName, burnCoins); err != nil {
				return err
			}
			if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins); err != nil {
				return err
			}
		}
	}

	// Validators portion remains in the fee collector; ensure non-negative invariant.
	for _, coin := range k.bankKeeper.GetAllBalances(ctx, feeCollectorAddr) {
		if coin.Amount.IsNegative() {
			return types.ErrInvalidParamsWrapf("fee collector negative balance for denom %s", coin.Denom)
		}
	}

	return nil
}

// helper for deterministic math (keeps sdk.Int copy local).
func floorPortion(total sdkmath.Int, bps uint32) sdkmath.Int {
	if total.IsZero() || bps == 0 {
		return sdkmath.NewInt(0)
	}
	return total.MulRaw(int64(bps)).QuoRaw(10000)
}

func (k Keeper) InitGenesis(ctx sdk.Context, gs types.GenesisState) error {
	return k.SetParams(ctx, gs.Params)
}

func (k Keeper) ExportGenesis(ctx sdk.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	return &types.GenesisState{Params: params}, nil
}

// paramsValueCodec stores Params using JSON to avoid proto requirements.
type paramsValueCodec struct{}

func (paramsValueCodec) Encode(value types.Params) ([]byte, error) {
	return json.Marshal(value)
}

func (paramsValueCodec) Decode(b []byte) (types.Params, error) {
	if len(b) == 0 {
		return types.Params{}, fmt.Errorf("empty params value")
	}
	var v types.Params
	return v, json.Unmarshal(b, &v)
}

func (paramsValueCodec) Stringify(value types.Params) string {
	return fmt.Sprintf("%+v", value)
}

func (paramsValueCodec) ValueType() string {
	return "feesplit-params"
}

func (paramsValueCodec) EncodeJSON(value types.Params) ([]byte, error) {
	return json.Marshal(value)
}

func (paramsValueCodec) DecodeJSON(b []byte) (types.Params, error) {
	return paramsValueCodec{}.Decode(b)
}
