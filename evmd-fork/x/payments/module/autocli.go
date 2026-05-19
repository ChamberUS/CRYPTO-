package payments

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/buynnex/iaos-evmd/x/payments/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Shows the parameters of the module",
				},
				{
					RpcMethod: "PaymentRequest",
					Use:       "payment-request [id]",
					Short:     "Get a payment request by id",
				},
				{
					RpcMethod: "PaymentRequestsByLoja",
					Use:       "payment-requests-by-loja [loja-id]",
					Short:     "List payment requests for a loja",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod: "CreatePaymentRequest",
					Use:       "create-payment-request [loja-id] [amount-microbyx]",
					Short:     "Create a new payment request",
				},
				{
					RpcMethod: "PayPaymentRequest",
					Use:       "pay-payment-request [request-id]",
					Short:     "Pay a pending payment request",
				},
			},
		},
	}
}
