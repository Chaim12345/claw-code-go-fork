# DeepSeek web API — per-message character limits

Measured live against the real chat endpoint via `StreamResponse` (the same
path the agent uses), with `deltaMode = true` to bypass the client-side
pre-flight cap (`provider.go:270`) so we measure the SERVER's enforcement.

## Results (probe run: 2026-06-07)

| Model   | Largest size accepted | Smallest size rejected | Approx server limit |
|---------|----------------------:|-----------------------:|--------------------:|
| default | 3,000,000 chars       | (none within tested)   | **≥ 750,000 tokens** |
| expert  | 193,753 chars         | 200,000 chars          | **48K – 50K tokens** |
| vision  | 2,906,253 chars       | 3,000,000 chars        | **726K – 750K tokens** |

- Probe: `internal/api/providers/deepseek/probe_limit_live_test.go` (build tag `live`)
- Run: `DEEPSEEK_TOKEN=… go test -tags=live -run TestLiveProbeActualInputLimit -v ./internal/api/providers/deepseek/`
- Total runtime: 83s (16s / 14s / 52s per model)
- Error message observed on reject: `stream error: Content is too long. Please shorten it and try again.`

## How the probe works

For each model, with a fresh `chat_session`:

1. **Phase 1 (ceiling probe)**: send a single message of `ceiling` chars.
   - Accepted → the model's limit is ≥ ceiling. No further probes.
   - Rejected → binary-search down between `floor` and `ceiling`.
2. **Phase 2 (binary search)**: split `[floor, ceiling]`, probe the midpoint.
   - Accepted → raise the floor to the midpoint.
   - Rejected → lower the ceiling to the midpoint.
   - Stop when `ceiling - floor > max(ceiling/20, 1000)` or after 12 attempts.
3. Report the final `lastOK` (largest accepted) and `hi` (smallest rejected).

Direction is large-to-small: the first probe is the ceiling, so we either
short-circuit (limit is at least the ceiling) or get a strong reject signal
that validates the harness before narrowing.

## Comparison with the current client caps

| Model   | Current `limitForModel` fallback | Server actually accepts | Slack |
|---------|---------------------------------:|------------------------:|------:|
| default | 60,000 tokens (240,000 chars)    | ≥ 750,000 tokens        | 12×+  |
| expert  | 38,000 tokens (152,000 chars)    | ~49,000 tokens          | 1.3×  |
| vision  | 600,000 tokens (2,400,000 chars) | ~738,000 tokens         | 1.2×  |

**Implication**: the pre-flight cap on `expert` (38K tokens, set in
`provider.go:104-106`) is *too tight* — the agent bails out and asks the
user to compact when the server would have accepted the prompt. The
default and vision caps are well below the server limit but harmless
because typical sessions never approach them.

`/api/v0/client/settings` is dead (returns `biz_code:2 INVALID_PARAM`),
so the cap is no longer auto-discoverable. The `limitForModel` fallback
in `provider.go:102-111` is the only source of truth at runtime.

## Caveats

- Single account, single point in time. Real limits may vary by account
  tier, server load, or model availability region.
- The probe uses a single user message; in production, the rendered
  prompt includes the system prompt + tool definitions + the full
  message history, so the *effective* per-message budget is lower.
- The error message ("Content is too long") doesn't disclose the
  server's actual cap — only the `input_character_limit` knob in
  `/settings` (now dead) ever did.
- The probe's accept signal is `message_start` (the server opened the
  stream). It does not wait for tokens to flow, so it's fast (≤90s
  per attempt) and unaffected by model latency.

## Server-error-driven recovery (no offline tokenizer)

The hard preflight reject at `provider.go:270-275` (the old `if modelCap
> 0 && !deltaMode && estimatedInput > modelCap { return … }`) was removed
because the local `chars/4` estimate is unreliable and the server's cap
is the only ground truth.

What replaces it:

1. **`StreamResponse` no longer rejects pre-flight.** It only emits a
   soft stderr warning when the local estimate exceeds
   `softPromptLimit` (28,000 tokens). Requests are sent as-is.
2. **Server's authoritative error triggers compaction.** When the
   server returns the transient `Content is too long. Please shorten it
   and try again.` error, the WebClient wraps it in a Go error. The
   conversation loop's `isPromptTooLargeError` (in
   `internal/runtime/conversation.go`) matches the substring
   `"content is too long"` (case-insensitive) and triggers `CompactNow`
   + retry.
3. **Retry budget raised from 1 → 3** in both `SendMessage` and
   `runOneTurnStreaming`. One compaction pass is often not enough to
   shrink below the cap; three gives the loop room to compact
   repeatedly until the rendered prompt fits.

### Why not an offline tokenizer?

We considered adding a Go BPE library (e.g.
`github.com/pkoukk/tiktoken-go`, `github.com/localit-io/tiktoken-go`,
`github.com/weaviate/tiktoken-go`, or
`github.com/tggo/goSentencePiece`) to count input tokens exactly.
Rejected because:

- **DeepSeek V3 has a custom tokenizer** (not GPT-4's `cl100k_base`,
  not a public SentencePiece model). Even the canonical V3 repo on
  Hugging Face does not ship a tokenizer config that matches what the
  web API uses server-side. We would still be estimating.
- **The server only returns `accumulated_token_usage` (output tokens)**
  in the SSE stream. There is no server-emitted input-token field we
  could parse, so we can't cross-check a local count against the
  server's view.
- **4MB of vocab data** to embed (`cl100k_base`) is non-trivial, and
  the accuracy gain (within ~10% of true count) does not change the
  design: the server is the contract, the local count is a hint.

We keep `EstimateTokens(s) = (len(s) + 3) / 4` as a sizing hint for the
soft warning and for `FitToBudget`, but the recovery loop trusts the
server's error, not the local estimate.

## Probe direction (large-to-small) rationale

Small-to-large (start at 100 chars, grow to the ceiling) was the original
direction. Large-to-small was chosen because:

- The first probe confirms the ceiling *fails* (or succeeds, in which
  case the limit is ≥ ceiling — useful information on its own).
- The first probe is the highest-risk one (biggest payload, biggest
  cost), so testing it first surfaces infrastructure issues (PoW,
  rate limit, missing headers) immediately rather than after several
  successful small probes.
- The 3× retry the WebClient does on "Content is too long" is only
  exercised by the first probe in large-to-small mode, so the total
  wasted work is bounded.
