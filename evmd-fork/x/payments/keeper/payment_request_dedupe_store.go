package keeper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/buynnex/iaos-evmd/x/payments/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) dedupeKey(lojaID uint64, amountMicro uint64, memo string) []byte {
	_, hash := fingerprintAndHash(lojaID, amountMicro, memo)

	key := make([]byte, 0, 8+1+len(hash))
	key = append(key, types.PaymentRequestDedupePrefix...)
	key = append(key, sdk.Uint64ToBigEndian(lojaID)...)
	key = append(key, hash[:]...)
	return key
}

func fingerprintAndHash(lojaID uint64, amountMicro uint64, memo string) (string, [32]byte) {
	fp := fmt.Sprintf("%d|%d|%s", lojaID, amountMicro, strings.TrimSpace(memo))
	return fp, sha256.Sum256([]byte(fp))
}

func traceIDFromCtx(ctx sdk.Context) string {
	tx := ctx.TxBytes()
	if len(tx) == 0 {
		return ""
	}
	sum := sha256.Sum256(tx)
	return fmt.Sprintf("%x", sum[:])
}

func (k Keeper) GetDedupeRequestID(ctx sdk.Context, lojaID, amountMicro uint64, memo string) (uint64, bool) {
	store := k.getStore(ctx)
	key := k.dedupeKey(lojaID, amountMicro, memo)
	bz := store.Get(key)
	if bz == nil {
		return 0, false
	}
	return binary.BigEndian.Uint64(bz), true
}

func (k Keeper) SetDedupeRequestID(ctx sdk.Context, lojaID, amountMicro uint64, memo string, requestID uint64) {
	store := k.getStore(ctx)
	key := k.dedupeKey(lojaID, amountMicro, memo)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, requestID)
	store.Set(key, bz)
}

func (k Keeper) DeleteDedupeRequestID(ctx sdk.Context, lojaID, amountMicro uint64, memo string) {
	store := k.getStore(ctx)
	key := k.dedupeKey(lojaID, amountMicro, memo)
	store.Delete(key)
}
