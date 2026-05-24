// Simple mock merchant webhook endpoint for E2E tests.
import { createServer } from "http";
import { createHmac } from "crypto";
import fs from "fs";

const PORT = Number(process.env.PORT || 4000);
const SECRET = process.env.MERCHANT_WEBHOOK_SECRET || process.env.SECRET || "devsecret";
const EVENTS_LOG_PATH = process.env.MOCK_EVENTS_LOG_PATH || "";
let remainingFails = Number(process.env.FAIL_FIRST_N || 0);
const seenKeys = new Set();

const log = (...args) => console.log("[MOCK]", ...args);

function appendEvent(event) {
  if (!EVENTS_LOG_PATH) return;
  try {
    fs.appendFileSync(EVENTS_LOG_PATH, `${JSON.stringify(event)}\n`);
  } catch (err) {
    log("failed to append events log", err);
  }
}

const server = createServer((req, res) => {
  if (req.method !== "POST" || req.url !== "/webhook") {
    res.statusCode = 404;
    res.end("not found");
    return;
  }

  const chunks = [];
  req.on("data", (c) => chunks.push(c));
  req.on("end", () => {
    const raw = Buffer.concat(chunks);
    const signature = req.headers["x-byx-signature"];
    const expected = createHmac("sha256", SECRET).update(raw).digest("hex");

    const idemKey = req.headers["x-byx-idempotency-key"];

    if (remainingFails > 0) {
      const attempt = Number(process.env.FAIL_FIRST_N || 0) - remainingFails + 1;
      log(`failing intentionally (${attempt}/${process.env.FAIL_FIRST_N})`);
      remainingFails -= 1;
      res.statusCode = 500;
      res.end("intentional failure");
      return;
    }

    if (idemKey && seenKeys.has(idemKey)) {
      log(`duplicate idempotency key ${idemKey}`);
      appendEvent({
        status: "duplicate",
        request_id: null,
        amount_ubyx: null,
        idempotency_key: idemKey,
        event_id: req.headers["x-byx-event-id"] || null,
      });
      res.statusCode = 200;
      res.end("duplicate ok");
      return;
    }

    if (signature !== expected) {
      log("invalid signature");
      appendEvent({
        status: "invalid_signature",
        request_id: null,
        amount_ubyx: null,
        idempotency_key: idemKey || null,
        event_id: req.headers["x-byx-event-id"] || null,
      });
      res.statusCode = 401;
      res.end("invalid signature");
      return;
    }

    let parsed;
    try {
      parsed = JSON.parse(raw.toString("utf8"));
    } catch (err) {
      log("invalid json body", err);
    }

    const reqId = parsed?.request_id ?? "?";
    const amount = parsed?.amount ?? parsed?.amount_ubyx ?? "?";
    if (idemKey) {
      seenKeys.add(idemKey);
    }

    log(`valid webhook request_id=${reqId} amount=${amount} idem=${idemKey ?? "none"}`);
    appendEvent({
      status: "ok",
      request_id: reqId,
      amount_ubyx: amount,
      idempotency_key: idemKey || null,
      event_id: req.headers["x-byx-event-id"] || null,
    });

    res.statusCode = 200;
    res.end("ok");
  });
});

server.listen(PORT, () => {
  log(`listening on :${PORT} (secret set, fail_first_n=${remainingFails})`);
});
