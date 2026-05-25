package payments

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/buynnex-corp/byx/x/payments/types"
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
				{
					RpcMethod:      "PaymentsQRCode",
					Use:            "payments-qr [request-id]",
					Short:          "Get compact QR payload JSON for a payment request",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "request_id"}},
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
					RpcMethod:      "CreatePaymentRequest",
					Use:            "create-payment-request [loja-id] [amount-ubyx]",
					Short:          "Create a new payment request",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "loja_id"}, {ProtoField: "amount_microbyx"}},
				},
				{
					RpcMethod:      "PayPaymentRequest",
					Use:            "pay-payment-request [request-id]",
					Short:          "Pay a pending payment request",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "request_id"}},
				},
				{
					RpcMethod:      "PayWithCertificate",
					Use:            "pay-with-certificate [request-id] [certificate-id]",
					Short:          "Pay and transfer certificate atomically",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "request_id"}, {ProtoField: "certificate_id"}},
				},
			},
		},
	}
}
