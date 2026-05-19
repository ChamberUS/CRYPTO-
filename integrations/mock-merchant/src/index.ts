import express from "express";
import crypto from "crypto";

const app = express();

// precisamos do RAW body para validar assinatura
app.use(express.json({
  verify: (req, _res, buf) => {
    // @ts-expect-error - anexando rawBody no request
    req.rawBody = buf;
  },
}));

const PORT = Number(process.env.MOCK_MERCHANT_PORT || "3001");
const SECRET = process.env.MERCHANT_WEBHOOK_SECRET || "dev_secret_123";

function timingSafeEqual(a: string, b: string) {
  const ab = Buffer.from(a);
  const bb = Buffer.from(b);
  if (ab.length !== bb.length) return false;
  return crypto.timingSafeEqual(ab, bb);
}

// Healthcheck
app.get("/health", (_req, res) => {
  res.status(200).json({ ok: true, service: "mock-merchant", port: PORT });
});

/**
 * Webhook endpoint esperado pelo relay: POST /webhook
 * Headers esperados (exemplo):
 *  - x-byx-event-id: <string>
 *  - x-byx-signature: <hex>  (HMAC-SHA256 do rawBody usando SECRET)
 */
app.post("/webhook", (req, res) => {
  const eventId = String(req.header("x-byx-event-id") || "");
  const sig = String(req.header("x-byx-signature") || "");

  // @ts-expect-error rawBody foi anexado no verify
  const rawBody: Buffer = req.rawBody || Buffer.from(JSON.stringify(req.body ?? {}));

  const computed = crypto.createHmac("sha256", SECRET).update(rawBody).digest("hex");
  const sigOk = sig.length > 0 && timingSafeEqual(sig, computed);

  const payload = req.body;

  const now = new Date().toISOString();
  console.log(`[MOCK-MERCHANT] ${now} received webhook`);
  console.log(`[MOCK-MERCHANT] eventId=${eventId || "(missing)"} sigOk=${sigOk}`);
  console.log(`[MOCK-MERCHANT] payload=`, payload);

  if (!sigOk) {
    return res.status(401).json({
      ok: false,
      error: "invalid_signature",
      eventId,
      computedPreview: computed.slice(0, 12) + "...",
    });
  }

  return res.status(200).json({ ok: true, received: true, eventId });
});

app.listen(PORT, () => {
  console.log(`[MOCK-MERCHANT] listening on http://127.0.0.1:${PORT}`);
  console.log(`[MOCK-MERCHANT] POST http://127.0.0.1:${PORT}/webhook`);
  console.log(`[MOCK-MERCHANT] GET  http://127.0.0.1:${PORT}/health`);
  console.log(`[MOCK-MERCHANT] SECRET (MERCHANT_WEBHOOK_SECRET) length=${SECRET.length}`);
});
