# Webhook IAOS v1 — Payment Settled

## Endpoint (merchant)
POST {MERCHANT_WEBHOOK_URL}

## Headers
- x-iaos-signature: sha256=<hex_hmac>
- x-iaos-event-id: <string>
- x-iaos-event-type: payment.settled.v1
- content-type: application/json

## Body (JSON)
{
  "event_id": "string",
  "event_type": "payment.settled.v1",
  "sent_at_unix": "int64",
  "data": {
    "request_id": "uint64",
    "loja_id": "uint64",
    "amount_microbyx": "string",
    "paid_at_unix": "int64",
    "payer": "string"
  }
}

## Assinatura
HMAC-SHA256 do corpo cru (bytes) usando MERCHANT_WEBHOOK_SECRET.
Header: x-iaos-signature: sha256=<hex>
