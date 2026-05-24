package app

import (
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	lojaskeeper "github.com/buynnex-corp/byx/x/lojas/keeper"
	lojastypes "github.com/buynnex-corp/byx/x/lojas/types"
	paymentskeeper "github.com/buynnex-corp/byx/x/payments/keeper"
	paymentstypes "github.com/buynnex-corp/byx/x/payments/types"
)

func TestPaymentsCreditsRealLojasMerchantSaldo(t *testing.T) {
	merchantAddr := sdk.AccAddress(bytesOf(0x01))
	payerAddr := sdk.AccAddress(bytesOf(0x02))
	payerInitialUbyx := sdk.NewInt64Coin(lojastypes.DenomBYX, 1_000_000)

	byxApp := New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, viper.New(), fauxMerkleModeOpt, baseapp.SetChainID(SimAppChainID))
	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)

	genesis, err := simtestutil.GenesisStateWithValSet(
		byxApp.AppCodec(),
		byxApp.DefaultGenesis(),
		valSet,
		[]authtypes.GenesisAccount{
			authtypes.NewBaseAccountWithAddress(merchantAddr),
			authtypes.NewBaseAccountWithAddress(payerAddr),
		},
		banktypes.Balance{
			Address: payerAddr.String(),
			Coins:   sdk.NewCoins(payerInitialUbyx),
		},
	)
	require.NoError(t, err)

	stateBytes, err := json.Marshal(genesis)
	require.NoError(t, err)

	_, err = byxApp.InitChain(&abci.RequestInitChain{
		AppStateBytes: stateBytes,
		ChainId:       SimAppChainID,
		Time:          time.Unix(1_700_000_000, 0).UTC(),
	})
	require.NoError(t, err)

	ctx := byxApp.NewContextLegacy(false, cmtproto.Header{
		Height: 1,
		Time:   time.Unix(1_700_000_001, 0).UTC(),
	})

	merchantCreator := merchantAddr.String()
	payer := payerAddr.String()
	lojasMsgServer := lojaskeeper.NewMsgServerImpl(byxApp.LojasKeeper)
	paymentsMsgServer := paymentskeeper.NewMsgServerImpl(byxApp.PaymentsKeeper)

	createMerchantResp, err := lojasMsgServer.CreateMerchant(ctx, &lojastypes.MsgCreateMerchant{
		Creator: merchantCreator,
		Nome:    "Loja Integracao",
	})
	require.NoError(t, err)

	createdMerchant, err := byxApp.LojasKeeper.Merchant(ctx, &lojastypes.QueryGetMerchantRequest{Id: createMerchantResp.Id})
	require.NoError(t, err)
	require.Equal(t, "0", createdMerchant.Merchant.Saldo)

	initialSupply := byxApp.BankKeeper.GetSupply(ctx, lojastypes.DenomBYX)

	createPaymentResp, err := paymentsMsgServer.CreatePaymentRequest(ctx, &paymentstypes.MsgCreatePaymentRequest{
		Creator:    merchantCreator,
		LojaId:     createMerchantResp.Id,
		AmountUbyx: 2_000,
		Memo:       "integration-saldo",
	})
	require.NoError(t, err)

	_, err = paymentsMsgServer.PayPaymentRequest(ctx, &paymentstypes.MsgPayPaymentRequest{
		Creator:   payer,
		RequestId: createPaymentResp.Id,
	})
	require.NoError(t, err)

	paidPayment, found := byxApp.PaymentsKeeper.GetPaymentRequest(ctx, createPaymentResp.Id)
	require.True(t, found)
	require.Equal(t, paymentstypes.PaymentStatus_PAYMENT_STATUS_PAID, paidPayment.Status)

	queriedMerchant, err := byxApp.LojasKeeper.Merchant(ctx, &lojastypes.QueryGetMerchantRequest{Id: createMerchantResp.Id})
	require.NoError(t, err)
	require.Equal(t, "2000", queriedMerchant.Merchant.Saldo)

	require.Equal(t, initialSupply, byxApp.BankKeeper.GetSupply(ctx, lojastypes.DenomBYX))
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 998_000), byxApp.BankKeeper.GetBalance(ctx, payerAddr, lojastypes.DenomBYX))
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 2_000), byxApp.BankKeeper.GetBalance(ctx, merchantAddr, lojastypes.DenomBYX))
}

func TestRegistrarVendaCashbackUsesModuleReserve(t *testing.T) {
	merchantAddr := sdk.AccAddress(bytesOf(0x11))
	customerAddr := sdk.AccAddress(bytesOf(0x12))
	moduleAddr := authtypes.NewModuleAddress(lojastypes.ModuleName)
	initialReserveUbyx := sdk.NewInt64Coin(lojastypes.DenomBYX, 80_000)

	byxApp, ctx := initTestApp(t,
		[]authtypes.GenesisAccount{
			authtypes.NewBaseAccountWithAddress(merchantAddr),
			authtypes.NewBaseAccountWithAddress(customerAddr),
		},
		[]banktypes.Balance{
			{Address: moduleAddr.String(), Coins: sdk.NewCoins(initialReserveUbyx)},
		},
	)

	merchantCreator := merchantAddr.String()
	customer := customerAddr.String()
	lojasMsgServer := lojaskeeper.NewMsgServerImpl(byxApp.LojasKeeper)

	_, err := lojasMsgServer.CreateMerchant(ctx, &lojastypes.MsgCreateMerchant{
		Creator: merchantCreator,
		Nome:    "Loja Cashback",
	})
	require.NoError(t, err)

	initialSupply := byxApp.BankKeeper.GetSupply(ctx, lojastypes.DenomBYX)
	initialCustomerBalance := byxApp.BankKeeper.GetBalance(ctx, customerAddr, lojastypes.DenomBYX)

	resp, err := lojasMsgServer.RegistrarVenda(ctx, &lojastypes.MsgRegistrarVenda{
		Creator:         merchantCreator,
		LojaId:          "1",
		ValorEmCentavos: 2_000,
		Cliente:         customer,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(50_000), resp.CashbackUbyx)

	require.Equal(t, initialSupply, byxApp.BankKeeper.GetSupply(ctx, lojastypes.DenomBYX))
	require.Equal(t, initialCustomerBalance.AddAmount(sdkmath.NewInt(50_000)), byxApp.BankKeeper.GetBalance(ctx, customerAddr, lojastypes.DenomBYX))
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 30_000), byxApp.BankKeeper.GetBalance(ctx, moduleAddr, lojastypes.DenomBYX))

	sale, err := byxApp.LojasKeeper.Sales.Get(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, uint64(50_000), sale.CashbackUbyx)
}

func TestRegistrarVendaCashbackFailsWithoutReserve(t *testing.T) {
	merchantAddr := sdk.AccAddress(bytesOf(0x21))
	customerAddr := sdk.AccAddress(bytesOf(0x22))
	moduleAddr := authtypes.NewModuleAddress(lojastypes.ModuleName)

	byxApp, ctx := initTestApp(t,
		[]authtypes.GenesisAccount{
			authtypes.NewBaseAccountWithAddress(merchantAddr),
			authtypes.NewBaseAccountWithAddress(customerAddr),
		},
		nil,
	)

	merchantCreator := merchantAddr.String()
	customer := customerAddr.String()
	lojasMsgServer := lojaskeeper.NewMsgServerImpl(byxApp.LojasKeeper)

	_, err := lojasMsgServer.CreateMerchant(ctx, &lojastypes.MsgCreateMerchant{
		Creator: merchantCreator,
		Nome:    "Loja Sem Reserva",
	})
	require.NoError(t, err)

	_, err = lojasMsgServer.RegistrarVenda(ctx, &lojastypes.MsgRegistrarVenda{
		Creator:         merchantCreator,
		LojaId:          "1",
		ValorEmCentavos: 2_000,
		Cliente:         customer,
	})
	require.Error(t, err)

	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 0), byxApp.BankKeeper.GetBalance(ctx, customerAddr, lojastypes.DenomBYX))
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 0), byxApp.BankKeeper.GetBalance(ctx, moduleAddr, lojastypes.DenomBYX))
	_, err = byxApp.LojasKeeper.Sales.Get(ctx, 1)
	require.Error(t, err)
	_, err = byxApp.LojasKeeper.SalesCount.Get(ctx)
	require.Error(t, err)
}

func TestLojasSupplyCapBlocksReserveTransfers(t *testing.T) {
	merchantAddr := sdk.AccAddress(bytesOf(0x31))
	customerAddr := sdk.AccAddress(bytesOf(0x32))
	moduleAddr := authtypes.NewModuleAddress(lojastypes.ModuleName)

	byxApp, ctx := initTestApp(t,
		[]authtypes.GenesisAccount{
			authtypes.NewBaseAccountWithAddress(merchantAddr),
			authtypes.NewBaseAccountWithAddress(customerAddr),
		},
		[]banktypes.Balance{
			{Address: moduleAddr.String(), Coins: sdk.NewCoins(sdk.NewInt64Coin(lojastypes.DenomBYX, 80_000))},
		},
	)

	currentSupply := byxApp.BankKeeper.GetSupply(ctx, lojastypes.DenomBYX).Amount
	capInt := sdkmath.NewInt(lojastypes.SupplyCapBYX)
	if currentSupply.LTE(capInt) {
		excess := capInt.Sub(currentSupply).AddRaw(1)
		require.NoError(t, byxApp.BankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(lojastypes.DenomBYX, excess))))
	}

	require.Error(t, byxApp.LojasKeeper.AssertSupplyCap(ctx))

	merchantCreator := merchantAddr.String()
	customer := customerAddr.String()
	lojasMsgServer := lojaskeeper.NewMsgServerImpl(byxApp.LojasKeeper)

	_, err := lojasMsgServer.CreateMerchant(ctx, &lojastypes.MsgCreateMerchant{
		Creator: merchantCreator,
		Nome:    "Loja Cap",
	})
	require.NoError(t, err)

	_, err = lojasMsgServer.RegistrarVenda(ctx, &lojastypes.MsgRegistrarVenda{
		Creator:         merchantCreator,
		LojaId:          "1",
		ValorEmCentavos: 2_000,
		Cliente:         customer,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "supply cap exceeded")
}

func TestFaucetTransfersUbyxFromReserve(t *testing.T) {
	adminAddr := sdk.AccAddress(bytesOf(0x41))
	moduleAddr := authtypes.NewModuleAddress(lojastypes.ModuleName)
	initialReserveUbyx := sdk.NewInt64Coin(lojastypes.DenomBYX, 3_000_000)

	byxApp, ctx := initTestApp(t,
		[]authtypes.GenesisAccount{
			authtypes.NewBaseAccountWithAddress(adminAddr),
		},
		[]banktypes.Balance{
			{Address: moduleAddr.String(), Coins: sdk.NewCoins(initialReserveUbyx)},
		},
	)

	lojasMsgServer := lojaskeeper.NewMsgServerImpl(byxApp.LojasKeeper)
	admin := adminAddr.String()

	createMerchantResp, err := lojasMsgServer.CreateMerchant(ctx, &lojastypes.MsgCreateMerchant{
		Creator: admin,
		Nome:    "Loja Faucet",
	})
	require.NoError(t, err)

	params, err := byxApp.LojasKeeper.ParamsStore.Get(ctx)
	require.NoError(t, err)
	params.FaucetEnabled = true
	params.FaucetAdmin = admin
	params.FaucetMaxPerTx = "3000000"
	require.NoError(t, byxApp.LojasKeeper.ParamsStore.Set(ctx, params))

	_, err = lojasMsgServer.Faucet(ctx, &lojastypes.MsgFaucet{
		Creator:   admin,
		LojistaId: "1",
		Amount:    "2000000",
	})
	require.NoError(t, err)

	merchantQuery, err := byxApp.LojasKeeper.Merchant(ctx, &lojastypes.QueryGetMerchantRequest{Id: createMerchantResp.Id})
	require.NoError(t, err)
	require.Equal(t, "2000000", merchantQuery.Merchant.Saldo)
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 1_000_000), byxApp.BankKeeper.GetBalance(ctx, moduleAddr, lojastypes.DenomBYX))
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 2_000_000), byxApp.BankKeeper.GetBalance(ctx, adminAddr, lojastypes.DenomBYX))
}

func TestRegistrarVendaCashbackRespectsUbyxCap(t *testing.T) {
	merchantAddr := sdk.AccAddress(bytesOf(0x51))
	customerAddr := sdk.AccAddress(bytesOf(0x52))
	moduleAddr := authtypes.NewModuleAddress(lojastypes.ModuleName)
	initialReserveUbyx := sdk.NewInt64Coin(lojastypes.DenomBYX, 10_000_000)

	byxApp, ctx := initTestApp(t,
		[]authtypes.GenesisAccount{
			authtypes.NewBaseAccountWithAddress(merchantAddr),
			authtypes.NewBaseAccountWithAddress(customerAddr),
		},
		[]banktypes.Balance{
			{Address: moduleAddr.String(), Coins: sdk.NewCoins(initialReserveUbyx)},
		},
	)

	merchantCreator := merchantAddr.String()
	customer := customerAddr.String()
	lojasMsgServer := lojaskeeper.NewMsgServerImpl(byxApp.LojasKeeper)

	_, err := lojasMsgServer.CreateMerchant(ctx, &lojastypes.MsgCreateMerchant{
		Creator: merchantCreator,
		Nome:    "Loja Cap Cashback",
	})
	require.NoError(t, err)

	params, err := byxApp.LojasKeeper.ParamsStore.Get(ctx)
	require.NoError(t, err)
	params.CashbackRateUbyxPerReal = 100_000
	params.MaxCashbackUbyxPorVenda = 25_000
	require.NoError(t, byxApp.LojasKeeper.ParamsStore.Set(ctx, params))

	resp, err := lojasMsgServer.RegistrarVenda(ctx, &lojastypes.MsgRegistrarVenda{
		Creator:         merchantCreator,
		LojaId:          "1",
		ValorEmCentavos: 2_000,
		Cliente:         customer,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(25_000), resp.CashbackUbyx)
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 25_000), byxApp.BankKeeper.GetBalance(ctx, customerAddr, lojastypes.DenomBYX))
	require.Equal(t, sdk.NewInt64Coin(lojastypes.DenomBYX, 9_975_000), byxApp.BankKeeper.GetBalance(ctx, moduleAddr, lojastypes.DenomBYX))
}

func initTestApp(t *testing.T, accounts []authtypes.GenesisAccount, balances []banktypes.Balance) (*App, sdk.Context) {
	t.Helper()

	byxApp := New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, viper.New(), fauxMerkleModeOpt, baseapp.SetChainID(SimAppChainID))
	valSet, err := simtestutil.CreateRandomValidatorSet()
	require.NoError(t, err)

	genesis, err := simtestutil.GenesisStateWithValSet(
		byxApp.AppCodec(),
		byxApp.DefaultGenesis(),
		valSet,
		accounts,
		balances...,
	)
	require.NoError(t, err)

	stateBytes, err := json.Marshal(genesis)
	require.NoError(t, err)

	_, err = byxApp.InitChain(&abci.RequestInitChain{
		AppStateBytes: stateBytes,
		ChainId:       SimAppChainID,
		Time:          time.Unix(1_700_000_000, 0).UTC(),
	})
	require.NoError(t, err)

	ctx := byxApp.NewContextLegacy(false, cmtproto.Header{
		Height: 1,
		Time:   time.Unix(1_700_000_001, 0).UTC(),
	})

	return byxApp, ctx
}

func bytesOf(value byte) []byte {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = value
	}
	return bz
}
