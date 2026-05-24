package keeper

import (
	"github.com/buynnex-corp/byx/x/payments/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensureCurrentStatus mutates pr when it is pending but already expired at the current block time.
// It persists the change to avoid returning stale PENDING states.
func (k Keeper) ensureCurrentStatus(ctx sdk.Context, pr *types.PaymentRequest) error {
	if pr == nil {
		return nil
	}
	if pr.Status != types.PaymentStatus_PAYMENT_STATUS_PENDING {
		return nil
	}
	if !pr.IsExpired(ctx.BlockTime().UTC()) {
		return nil
	}

	pr.Status = types.PaymentStatus_PAYMENT_STATUS_EXPIRED
	k.SetPaymentRequest(ctx, *pr)
	_ = k.PaymentRequests.Set(ctx, pr.Id, *pr)
	k.DeleteDedupeRequestID(ctx, pr.LojaId, pr.AmountUbyx, pr.Memo)
	return nil
}
