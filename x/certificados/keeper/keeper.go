package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

var _ types.QueryServer = (*Keeper)(nil)

type Keeper struct {
	types.UnimplementedQueryServer
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	authority    []byte

	Schema              collections.Schema
	ParamsStore         collections.Item[types.Params]
	Certificates        collections.Map[uint64, types.Certificate]
	CertificateCount    collections.Item[uint64]
	ByOwner             collections.KeySet[collections.Pair[string, uint64]]
	ByMerchant          collections.KeySet[collections.Pair[uint64, uint64]]
	BySerialHash        collections.KeySet[collections.Pair[string, uint64]]
	ServiceRecords      collections.Map[collections.Pair[uint64, uint64], types.ServiceRecord]
	ServiceRecordsCount collections.Map[uint64, uint64]
	bankKeeper          types.BankKeeper
	lojasKeeper         types.LojasKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	bk types.BankKeeper,
	lk types.LojasKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		bankKeeper:   bk,
		lojasKeeper:  lk,

		ParamsStore:         collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Certificates:        collections.NewMap(sb, types.CertificateKey, "certificates", collections.Uint64Key, codec.CollValue[types.Certificate](cdc)),
		CertificateCount:    collections.NewItem(sb, types.CertificateCountKey, "certificate_count", collections.Uint64Value),
		ByOwner:             collections.NewKeySet(sb, types.CertificateByOwnerIndex, "certificates_by_owner", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key)),
		ByMerchant:          collections.NewKeySet(sb, types.CertificateByMerchant, "certificates_by_merchant", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key)),
		BySerialHash:        collections.NewKeySet(sb, types.CertificateBySerial, "certificates_by_serial", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key)),
		ServiceRecords:      collections.NewMap(sb, types.ServiceRecordKey, "service_records", collections.PairKeyCodec(collections.Uint64Key, collections.Uint64Key), codec.CollValue[types.ServiceRecord](cdc)),
		ServiceRecordsCount: collections.NewMap(sb, types.ServiceCountKey, "service_records_count", collections.Uint64Key, collections.Uint64Value),
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
