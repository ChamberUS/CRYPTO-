import { createHmac } from "crypto";
import { setTimeout as delay } from "timers/promises";
import fs from "fs/promises";
import path from "path";
import { fileURLToPath } from "url";

type PaymentRequest = {
  id: number;
  loja_id?: number;
  lojaId?: number;
  amount_ubyx?: number;
  amountUbyx?: number;
  payer?: string;
  paid_at_unix?: number;
  paidAtUnix?: number;
  status?: string | number;
};

type StoredAttempt = {
  eventId: string;
  attempts?: number;
  sentAtUnix?: number;
  nextRetryAt?: number;
  lastError?: string;
  lastAttemptAt?: number;
  nextRetryAtUnix?: number; // backward compat
  payload?: WebhookPayload; // backward compat
};

type State = {
  sent: Record<string, { eventId: string; at: number }>;
  failures: Record<
    string,
    {
      count: number;
      lastError?: string;
      nextRetryAt?: number;
      lastAttemptAt?: number;
      lastEventId?: string;
      payload?: any;
    }
  >;
  deadLetters: Array<{
    requestId: string;
    lojaId: string;
    eventId?: string;
    payload?: any;
    lastError?: string;
    failedAt: number;
    attempts: number;
    nextAction?: "manual_replay";
  }>;
};

type WebhookPayload = {
  request_id: number;
  loja_id?: number;
  amount_ubyx?: number;
  payer?: string;
  paid_at_unix: number;
  event_id: string;
  trace_id?: string;
};

const REST_ENDPOINT = process.env.REST_ENDPOINT || "";
const MERCHANT_WEBHOOK_URL = process.env.MERCHANT_WEBHOOK_URL || "";
const MERCHANT_WEBHOOK_SECRET = process.env.MERCHANT_WEBHOOK_SECRET || "";
const LOJA_ID = process.env.LOJA_ID;
const POLL_MS = Number(process.env.POLL_MS || 2000);
const REQUEST_TIMEOUT_MS = 3000;
const LOG_PREFIX = "[BYX-WEBHOOK]";
const MAX_ATTEMPTS = Number(process.env.MAX_ATTEMPTS || 8);
const BASE_BACKOFF_MS = Number(process.env.BASE_BACKOFF_MS || 500);
const MAX_BACKOFF_MS = Number(process.env.MAX_BACKOFF_MS || 30000);
const REPLAY_DLQ = (process.env.REPLAY_DLQ || "").toLowerCase() === "1";
const DRY_RUN = (process.env.DRY_RUN || "").toLowerCase() === "true";
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const STATE_PATH = process.env.STATE_PATH || path.join(__dirname, "state.json");

if (!REST_ENDPOINT || !MERCHANT_WEBHOOK_URL || !MERCHANT_WEBHOOK_SECRET || !LOJA_ID) {
  console.error(
    `${LOG_PREFIX} Missing env vars. Required: REST_ENDPOINT, MERCHANT_WEBHOOK_URL, MERCHANT_WEBHOOK_SECRET, LOJA_ID`
  );
  process.exit(1);
}

const seenStatuses = new Map<number, string>();
const inflight = new Set<number>();
let state: State = { sent: {}, failures: {}, deadLetters: [] };

async function loadState() {
  try {
    const buf = await fs.readFile(STATE_PATH, "utf8");
    state = JSON.parse(buf) as State;
    if (!state.sent) state.sent = {};
    if (!state.failures) state.failures = {};
    if (!state.deadLetters) state.deadLetters = [];
  } catch (err: any) {
    if (err?.code === "ENOENT") {
      state = { sent: {}, failures: {}, deadLetters: [] };
      await saveState(); // bootstrap file
    } else {
      console.error(`${LOG_PREFIX} failed to read state:`, err);
    }
  }
  console.log(
    `${LOG_PREFIX} state loaded from ${STATE_PATH} sent=${Object.keys(state.sent).length} failures=${Object.keys(
      state.failures
    ).length} dlq=${state.deadLetters.length}`
  );
}

async function saveState() {
  await fs.writeFile(STATE_PATH, JSON.stringify(state, null, 2));
}

console.log(
  `${LOG_PREFIX} config REST=${REST_ENDPOINT} LOJA_ID=${LOJA_ID} WEBHOOK_URL=${MERCHANT_WEBHOOK_URL} POLL_MS=${POLL_MS} (secret set)`
);

function statusLabel(raw: string | number | undefined): string {
  if (typeof raw === "number") {
    if (raw === 2) return "PAID";
    if (raw === 3) return "EXPIRED";
    if (raw === 4) return "CANCELED";
    return "PENDING";
  }
  if (!raw) return "UNKNOWN";
  return raw.replace("PAYMENT_STATUS_", "");
}

function buildPayload(pr: PaymentRequest): WebhookPayload {
  const paidAt = pr.paid_at_unix ?? pr.paidAtUnix ?? Math.floor(Date.now() / 1000);
  const eventId = `${pr.id}:${paidAt}`;
  const traceId = pr.trace_id ?? pr.traceId ?? eventId;
  const payload: WebhookPayload = {
    request_id: pr.id,
    paid_at_unix: paidAt,
    event_id: eventId,
    trace_id: traceId,
  };
  const loja = pr.loja_id ?? pr.lojaId;
  if (loja !== undefined) payload.loja_id = loja;
  const amount = pr.amount_ubyx ?? pr.amountUbyx;
  if (amount !== undefined) payload.amount_ubyx = amount;
  if (pr.payer) payload.payer = pr.payer;
  return payload;
}

async function poll() {
  await processPendingFailures();
  try {
    const res = await fetch(
      `${REST_ENDPOINT}/byx/payments/v1/payment_requests/by_loja/${LOJA_ID}?pagination.limit=50`
    );
    if (!res.ok) {
      throw new Error(`REST ${res.status}`);
    }
    const data = await res.json();
    const requests: PaymentRequest[] = data.payment_requests || data.paymentRequests || [];

    for (const pr of requests) {
      const id = pr.id;
      const current = statusLabel(pr.status);
      const previous = seenStatuses.get(id);
      seenStatuses.set(id, current);

      if (current === "PAID") {
        const payload = buildPayload(pr);
        const existing = state.sent[id];
        if (existing && existing.eventId === payload.event_id) {
          continue; // already delivered with same event
        }
        if (!inflight.has(id)) {
          if (previous !== "PAID") {
            console.log(`${LOG_PREFIX} request ${id} detected as PAID (event ${payload.event_id})`);
          }
          void sendWebhookWithState(payload);
        }
      }
    }
  } catch (err) {
    console.error(`${LOG_PREFIX} poll error:`, err);
  } finally {
    setTimeout(poll, POLL_MS);
  }
}

function jitteredBackoff(baseMs: number): number {
  const jitter = Math.floor(Math.random() * 251); // 0..250ms
  return Math.min(baseMs + jitter, MAX_BACKOFF_MS);
}

async function sendWebhookWithState(payload: WebhookPayload) {
  const id = payload.request_id;
  if (inflight.has(id)) return;
  const sent = state.sent[String(id)];
  if (sent && sent.eventId === payload.event_id) {
    return;
  }
  inflight.add(id);
  const body = JSON.stringify({
    ...payload,
  });
  const signature = createHmac("sha256", MERCHANT_WEBHOOK_SECRET).update(body).digest("hex");

  let attempts = 0;

  const sendOnce = async () => {
    attempts += 1;
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);
    const attemptsLeft = MAX_ATTEMPTS - attempts;

    try {
      console.log(
        `${LOG_PREFIX} sending webhook for request ${id} (attempt ${attempts}/${MAX_ATTEMPTS}) event=${payload.event_id}`
      );

      if (DRY_RUN) {
        console.log(`${LOG_PREFIX} DRY_RUN enabled, skipping POST for request ${id}`);
        markSent(payload, attempts);
        inflight.delete(id);
        return;
      }

      const res = await fetch(MERCHANT_WEBHOOK_URL, {
        method: "POST",
        headers: {
          "content-type": "application/json",
          "X-BYX-Signature": signature,
          "X-BYX-Idempotency-Key": String(id),
          "X-BYX-Event-Id": payload.event_id,
        },
        body,
        signal: controller.signal,
      });

      if (res.ok) {
        console.log(`${LOG_PREFIX} delivered request=${id} loja=${payload.loja_id ?? "-"} event=${payload.event_id} trace=${payload.trace_id ?? ""} attempts=${attempts}`);
        markSent(payload, attempts);
        inflight.delete(id);
        return;
      }

      const reason = `status ${res.status}`;
      console.error(
        `${LOG_PREFIX} webhook responded ${reason} for request ${id} (remaining ${attemptsLeft})`
      );
      throw new Error(reason);
    } catch (err) {
      const reason = err instanceof Error ? err.message : String(err);
      const remaining = MAX_ATTEMPTS - attempts;
      if (attempts >= MAX_ATTEMPTS) {
        console.error(`${LOG_PREFIX} DLQ: request ${id} moved to deadLetters (attempts=${attempts}) error=${reason}`);
        moveToDeadLetters(payload, attempts, reason);
        inflight.delete(id);
        return;
      }
      const base = Math.min(BASE_BACKOFF_MS * 2 ** (attempts - 1), MAX_BACKOFF_MS);
      const wait = jitteredBackoff(base);
      console.error(
        `${LOG_PREFIX} send failed request=${id} event=${payload.event_id} attempt=${attempts}/${MAX_ATTEMPTS} remaining=${remaining} nextRetryAt=${Date.now() + wait}ms err=${reason}`
      );
      markFailure(payload, attempts, reason, Date.now() + wait);
      await delay(wait);
      await sendOnce();
    } finally {
      clearTimeout(timeoutId);
    }
  };

  await sendOnce();
}

function markSent(payload: WebhookPayload, attempts: number) {
  state.sent[String(payload.request_id)] = {
    eventId: payload.event_id,
    at: Date.now(),
  };
  delete state.failures[String(payload.request_id)];
  state.deadLetters = state.deadLetters.filter((d) => d.requestId !== String(payload.request_id));
  void saveState();
}

function markFailure(payload: WebhookPayload, attempts: number, reason: string, nextRetryAt: number) {
  state.failures[String(payload.request_id)] = {
    count: attempts,
    lastError: reason,
    nextRetryAt,
    lastAttemptAt: Date.now(),
    lastEventId: payload.event_id,
    payload,
  };
  void saveState();
}

function moveToDeadLetters(payload: WebhookPayload, attempts: number, reason: string) {
  state.deadLetters.push({
    requestId: String(payload.request_id),
    lojaId: String(payload.loja_id ?? ""),
    eventId: payload.event_id,
    payload,
    lastError: reason,
    failedAt: Date.now(),
    attempts,
    nextAction: "manual_replay",
  });
  delete state.failures[String(payload.request_id)];
  void saveState();
}

async function processPendingFailures() {
  const now = Date.now();
  for (const [idStr, failure] of Object.entries(state.failures)) {
    const id = Number(idStr);
    const nextRetryAt =
      failure.nextRetryAt !== undefined
        ? failure.nextRetryAt
        : failure.nextRetryAtUnix !== undefined
        ? (failure.nextRetryAtUnix as number) * 1000
        : undefined;
    if (!nextRetryAt || now < nextRetryAt) continue;
    if (inflight.has(id)) continue;
    const payload: WebhookPayload | undefined = (failure as any).payload || undefined;
    if (!payload) continue;
    console.log(
      `${LOG_PREFIX} retrying request=${id} event=${failure.lastEventId} attempt=${(failure.count || 0) + 1} nextRetryAt=${nextRetryAt}`
    );
    void sendWebhookWithState(payload);
  }

  if (REPLAY_DLQ && state.deadLetters.length > 0) {
    for (const dlq of [...state.deadLetters]) {
      const id = Number(dlq.requestId);
      if (inflight.has(id)) continue;
      if (!dlq.payload) continue;
      console.log(`${LOG_PREFIX} replaying DLQ request=${dlq.requestId} event=${dlq.eventId}`);
      void sendWebhookWithState(dlq.payload as WebhookPayload);
    }
  }
}

async function start() {
  await loadState();
  console.log(
    `${LOG_PREFIX} polling iniciado para loja ${LOJA_ID} (intervalo ${POLL_MS}ms) MAX_ATTEMPTS=${MAX_ATTEMPTS} DRY_RUN=${DRY_RUN} REPLAY_DLQ=${REPLAY_DLQ}`
  );
  const shutdown = async () => {
    console.log(`${LOG_PREFIX} shutting down, persisting state...`);
    await saveState();
    process.exit(0);
  };
  process.on("SIGINT", shutdown);
  process.on("SIGTERM", shutdown);
  poll();
}

start().catch((err) => {
  console.error(`${LOG_PREFIX} fatal error:`, err);
  process.exit(1);
});
