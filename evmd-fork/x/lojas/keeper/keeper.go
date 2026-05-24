package keeper

import (
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	math "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/buynnex/iaos-evmd/x/lojas/types"
)

var _ types.QueryServer = (*Keeper)(nil)

type Keeper struct {
	types.UnimplementedQueryServer
	storeService  corestore.KVStoreService
	cdc           codec.Codec
	addressCodec  address.Codec
	accountKeeper authkeeper.AccountKeeper
	bankKeeper    bankkeeper.Keeper
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema                collections.Schema
	ParamsStore           collections.Item[types.Params]
	Sales                 collections.Map[uint64, types.Sale]
	SalesCount            collections.Item[uint64]
	SalesByLojaIndex      collections.KeySet[collections.Pair[uint64, uint64]]
	DailyCashbackByLoja   collections.Map[collections.Pair[uint64, uint64], uint64]
	SalesCountStateByLoja collections.Map[uint64, types.SalesCountState]
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	ak authkeeper.AccountKeeper,
	bk bankkeeper.Keeper,

) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService:          storeService,
		cdc:                   cdc,
		addressCodec:          addressCodec,
		authority:             authority,
		accountKeeper:         ak,
		bankKeeper:            bk,
		ParamsStore:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Sales:                 collections.NewMap(sb, types.SalesKey, "sales", collections.Uint64Key, codec.CollValue[types.Sale](cdc)),
		SalesCount:            collections.NewItem(sb, types.SaleCountKey, "sales_count", collections.Uint64Value),
		SalesByLojaIndex:      collections.NewKeySet(sb, types.SalesByLojaIndexKey, "sales_by_loja", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key)),
		DailyCashbackByLoja:   collections.NewMap(sb, types.DailyCashbackByLojaKey, "daily_cashback_by_loja", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), collections.Uint64Value),
		SalesCountStateByLoja: collections.NewMap(sb, types.SalesCountStateByLojaKey, "sales_count_block_state", collections.Uint64Key, codec.CollValue[types.SalesCountState](cdc)),
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// nextSaleID increments and returns the next sale identifier.
func (k Keeper) nextSaleID(ctx sdk.Context) (uint64, error) {
	count, err := k.SalesCount.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			count = 0
		} else {
			return 0, err
		}
	}

	next := count + 1
	if err := k.SalesCount.Set(ctx, next); err != nil {
		return 0, err
	}
	return next, nil
}

// storeSale persists the sale and indexes it by loja.
func (k Keeper) storeSale(ctx sdk.Context, sale types.Sale) error {
	if err := k.Sales.Set(ctx, sale.Id, sale); err != nil {
		return err
	}
	return k.SalesByLojaIndex.Set(ctx, collections.Join(sale.LojaId, sale.Id))
}

// dayKey returns YYYYMMDD as uint64 in UTC.
func dayKey(t time.Time) uint64 {
	d := t.UTC()
	return uint64(d.Year())*10000 + uint64(d.Month())*100 + uint64(d.Day())
}

// CalculateCashbackFromCentavos calcula o cashback em ubyx a partir de um valor em centavos.
// Ele sempre retorna um sdk.Coin com denom types.DenomBYX.
func (k Keeper) CalculateCashbackFromCentavos(ctx sdk.Context, valorEmCentavos int64) sdk.Coin {
	// Se o valor em centavos for inválido ou zero/negativo, não há cashback.
	if valorEmCentavos <= 0 {
		return sdk.NewCoin(types.DenomBYX, math.NewInt(0))
	}

	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		// Em caso de erro ao ler os parâmetros, é mais seguro retornar 0.
		return sdk.NewCoin(types.DenomBYX, math.NewInt(0))
	}

	if params.CashbackRateUbyxPerReal == 0 {
		// Rate 0 significa que o módulo está configurado para não dar cashback.
		return sdk.NewCoin(types.DenomBYX, math.NewInt(0))
	}

	// Convertemos centavos para "reais" inteiros (divisão inteira).
	valorEmReais := valorEmCentavos / 100
	if valorEmReais <= 0 {
		// Valores menores que 100 centavos resultam em 0 reais => cashback 0.
		return sdk.NewCoin(types.DenomBYX, math.NewInt(0))
	}

	// cashbackMicro = valorEmReais * rate (ubyx por real)
	rate := int64(params.CashbackRateUbyxPerReal)
	cashbackMicro := int64(valorEmReais) * rate
	if cashbackMicro <= 0 {
		return sdk.NewCoin(types.DenomBYX, math.NewInt(0))
	}

	return sdk.NewCoin(types.DenomBYX, math.NewInt(cashbackMicro))
}

// MintBYXTo faz o mint de uma quantidade positiva de BYX para uma conta.
// Ele utiliza a module account do módulo "lojas" (types.ModuleName) como origem do mint.
func (k Keeper) MintBYXTo(ctx sdk.Context, addr sdk.AccAddress, amount math.Int) error {
	if !amount.IsPositive() {
		return fmt.Errorf("amount must be positive")
	}

	coin := sdk.NewCoin(types.DenomBYX, amount)
	coins := sdk.NewCoins(coin)

	// Mint na module account do módulo "lojas".
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}

	// Envia da module account para a conta de destino.
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, addr, coins); err != nil {
		return err
	}

	// Evento para facilitar indexação e debug.
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"lojas_mint_byx",
			sdk.NewAttribute("to", addr.String()),
			sdk.NewAttribute("amount", amount.String()),
		),
	)

	return nil
}
