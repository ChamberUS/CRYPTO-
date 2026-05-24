package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/buynnex-corp/byx/x/payments/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema                collections.Schema
	Params                collections.Item[types.Params]
	PaymentRequests       collections.Map[uint64, types.PaymentRequest]
	PaymentRequestSeq     collections.Sequence
	PaymentRequestsByLoja collections.KeySet[collections.Pair[uint64, uint64]]

	bankKeeper         types.BankKeeper
	lojasKeeper        types.LojasKeeper
	certificadosKeeper types.CertificadosKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	bk types.BankKeeper,
	lk types.LojasKeeper,
	ck types.CertificadosKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService:       storeService,
		cdc:                cdc,
		addressCodec:       addressCodec,
		authority:          authority,
		bankKeeper:         bk,
		lojasKeeper:        lk,
		certificadosKeeper: ck,

		Params:                collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		PaymentRequests:       collections.NewMap(sb, types.PaymentRequestKey, "payment_request", collections.Uint64Key, codec.CollValue[types.PaymentRequest](cdc)),
		PaymentRequestSeq:     collections.NewSequence(sb, types.PaymentRequestCountKey, "payment_request_seq"),
		PaymentRequestsByLoja: collections.NewKeySet(sb, types.PaymentRequestsByLojaPK, "payment_requests_by_loja", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key)),
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
