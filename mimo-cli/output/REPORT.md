# MiMo Studio Web App - Reverse Engineering Report

**Date**: 2026-06-01  
**Target**: https://aistudio.xiaomimimo.com  
**Status**: Complete  
**Analyst**: OpenCode + Claude

---

## Executive Summary

MiMo Studio is Xiaomi's full-featured AI chat platform, powered by their MiMo-2.5 series models. The web app is a React SPA that communicates with a WebSocket-based backend via a `/ws/proxy` gateway. The HTTP `/open-apis/bot/chat` endpoint serves legacy chat functionality, while the modern chat system runs entirely over WebSocket with an RPC pattern.

**Key finding**: Tool calling is **server-side only**. The client receives pre-executed tool call results via WebSocket stream events, but has no client-side MCP/tool configuration. The tool calling infrastructure is hidden behind the `mimo-agent` service, which requires gated provisioning.

---

## Architecture

### Frontend Stack
| Component | Technology |
|-----------|-----------|
| Framework | React (JSX) with webpack 5 |
| State | MobX (makeAutoObservable, runInAction) |
| UI Library | Arco Design (ByteDance) |
| Schema | Zod (validation) |
| Markdown | remark/rehype pipeline |
| Voice | LiveKit (WebRTC) real-time audio |
| Audio | FFmpeg WASM for client-side processing |
| i18n | zh-CN / en locale support |

### Backend Communication
| Channel | Usage |
|---------|-------|
| WebSocket `/ws/proxy` | Primary chat (send, history, abort) |
| REST `/open-apis/*` | Config, auth, files, TTS, sharing |
| LiveKit | Real-time voice conversations |
| Static CDN | Assets, policies, TTS audio |

---

## All API Endpoints (35 total)

### Chat & Agent (REST)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/open-apis/bot/chat` | Legacy HTTP chat (SSE streaming) |
| GET | `/open-apis/bot/config` | Get model registry + features |
| POST | `/open-apis/bot/chat/genUploadInfo` | Generate file upload URL |
| POST | `/open-apis/bot/chat/genNewUrl` | Refresh expired upload URL |
| POST | `/open-apis/bot/chat/parse` | Parse uploaded file |
| GET | `/open-apis/resource/config` | Resource/model config |

### Chat & Agent (WebSocket RPC)
| Method | Request | Purpose |
|--------|---------|---------|
| `chat.send` | `{sessionKey, message, deliver, idempotencyKey}` | Send message |
| `chat.abort` | `{sessionKey, runId}` | Abort generation |
| `chat.history` | `{sessionKey, limit}` | Load history |
| `sessions.list` | `{includeGlobal, limit}` | List all sessions |

### Conversation Management (REST)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/open-apis/chat/conversation/list` | List conversations |
| POST | `/open-apis/chat/conversation/save` | Save/update conversation |
| POST | `/open-apis/chat/conversation/genTitle` | Auto-generate title |
| POST | `/open-apis/chat/dialog/list` | List messages (paginated) |
| POST | `/open-apis/chat/dialog/feedback` | Submit message feedback |

### User & Auth (REST)
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/open-apis/user/mi/get` | Get Xiaomi user info |
| POST | `/open-apis/user/mi/logout` | Logout |
| POST | `/open-apis/v1/genLoginUrl` | Generate login redirect |
| GET | `/open-apis/agreement` | Check agreement status |
| POST | `/open-apis/agreement` | Accept agreement |

### Claw Agent (REST)
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/open-apis/user/mimo-claw/status` | Get agent resource status |
| POST | `/open-apis/user/mimo-claw/create` | Create agent resource |
| POST | `/open-apis/user/mimo-claw/destroy` | Destroy agent resource |
| POST | `/open-apis/user/mimo-claw/restart` | Restart agent |
| POST | `/open-apis/user/mimo-claw/repair` | Repair agent |
| POST | `/open-apis/agreement/user/mimo-claw` | Accept claw disclaimer |

### TTS (REST)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/open-apis/tts/v2/generate` | Generate TTS audio |
| GET | `/open-apis/tts/generateStatus` | Poll generation status |
| POST | `/open-apis/tts/download` | Download audio |
| POST | `/open-apis/tts/generateStyle` | Generate voice style |
| POST | `/open-apis/tts/generateSpeechText` | Generate speech text |

### Voice (REST)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/open-apis/chat/liveKit/getRtcConfig` | Get LiveKit token + config |

### LiveKit RPC
| Method | Purpose |
|--------|---------|
| `mimo.update_metadata` | Send voice config to agent |
| `mimo.cancel_turn` | Cancel voice turn |
| `mimo.custom_voice` | Upload custom voice |
| `lk.retract_message` | Recall voice message |

### File Management (REST)
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/open-apis/host-files/list` | List files on agent |
| POST | `/open-apis/host-files/preview` | Preview file |
| POST | `/open-apis/host-files/download` | Download file |

### Other (REST)
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/open-apis/share/createShare` | Create shareable link |
| POST | `/open-apis/audit/submit` | Submit content audit |
| POST | `/open-apis/contact` | Contact/feedback |

---

## WebSocket Architecture

### Connection Flow
```
1. GET /open-apis/user/ws/ticket → { ticket: "..." }
2. WS connect: wss://<host>/ws/proxy?ticket=<ticket>
3. Send RPC: { method: "chat.send", data: {sessionKey, message} }
4. Receive stream: { stream: "lifecycle", data: {phase: "start"} }
5. Receive stream: { stream: "chat", data: {state: "delta", content: "..."} }
6. Receive stream: { stream: "tool", data: {name, args, output} }
7. Receive stream: { stream: "chat", data: {state: "final", ...} }
```

### Stream Event Types
| Event | State | Purpose |
|-------|-------|---------|
| `lifecycle` | `start`, `error`, `end` | Connection lifecycle |
| `chat` | `delta` | Streaming text chunk |
| `chat` | `final` | Stream complete (includes usage stats) |
| `chat` | `aborted` | User cancelled |
| `chat` | `error` | Generation error |
| `chat` | `sensitive_query` | Content filtered |
| `tool` | — | Tool call event (name, args, output) |

---

## Tool Calling Analysis

### What Was Found
- **Client receives tool calls**: Stream events include `{type: "toolcalls", calls: [{id, name, args, phase, output}]}`
- **No client-side configuration**: No MCP servers, no tool definitions, no function schemas in client code
- **Tool types recognized**: `toolcall`, `tool_use`, `function_call`
- **Agent identity**: `mimo-agent` (LiveKit destination identity)

### What Was NOT Found
- No client-side MCP/tool configuration UI
- No `mcpServers` field in request bodies
- No `functions` or `tools` parameter support
- No `parallel_tool_calls` support
- No tool calling capability exposed via HTTP API

### Conclusion
Tool calling is **fully server-side**. The `mimo-agent` service handles tool execution remotely. The client only receives pre-executed results. To use tool calling programmatically, you need:
1. **Provisioned Claw agent** (gated access via `/open-apis/user/mimo-claw/*`)
2. **WebSocket connection** (not HTTP REST)
3. **Proper session management** (sessionKey + idempotencyKey)

---

## Model Registry

### Available Models (from `/open-apis/bot/config`)
| Model ID | Display | Page Type | Default | Features |
|----------|---------|-----------|---------|----------|
| `mimo-v2-flash` | MiMo V2 Flash | chat | — | web search, thinking |
| `mimo-v2-flash-studio` | MiMo V2 Flash Studio | chat | — | web search, thinking |
| `mimo-v2-pro` | MiMo V2 Pro | chat | — | web search, thinking |
| `mimo-v2.5` | MiMo V2.5 | chat | — | web search, thinking |
| `mimo-v2.5-pro` | MiMo V2.5 Pro | chat | ✓ (default) | web search, thinking |
| `clawm-alpha` | MiMo Claw Agent (Pro) | claw | — | **GATED** |
| `clawl-alpha` | MiMo Claw Agent (Lite) | claw | — | **GATED** |
| `mimo-v2-tts` | MiMo V2 TTS | tts | — | 8 voice options |
| `mimo-v2-tts-vc` | MiMo V2 TTS-VC | tts-vc | — | Voice conversion |
| `mimo-v2-tts-vd` | MiMo V2 TTS-VD | tts-vd | — | Voice design |

### Hidden Models (not in UI)
- `clawm-alpha` — Claw Agent Pro (resource-intensive, gated)
- `clawl-alpha` — Claw Agent Lite (resource-lighter, gated)
- `mimo-v2-flash-studio` — Studio variant of Flash
- `mimo-v2-tts-vc` — Voice conversion variant
- `mimo-v2-tts-vd` — Voice design variant

---

## Auth System

### Cookie-Based Session
| Cookie | Domain | Purpose |
|--------|--------|---------|
| `serviceToken` | aistudio.xiaomimimo.com | Primary session token |
| `userId` | aistudio.xiaomimimo.com | User identifier |
| `xiaomichatbot_ph` | aistudio.xiaomimimo.com | Chat session key |

### Request Headers
```
Accept-Language: zh-CN
x-timeZone: America/New_York
credentials: same-origin
```

### Error Handling
| HTTP Code | Action |
|-----------|--------|
| 401/302 | Redirect to login |
| 451 | Permanent ban |
| 461 | Temporary ban (chat/bot/tts) |
| 429/503 | Show service busy overlay |

---

## Static Assets Inventory

### From aistudio.xiaomimimo.com (18 files)
- 14 JS bundles (6.1MB total)
- 4 CSS files (2.2MB total)
- 9 images (PNG/WebP)

### From aistudio-cdn.xiaomimimo.com (27 files)
- FFmpeg WASM: 3.5MB (core.js + core.wasm)
- TTS audio samples: 8 voices × WAV (2.5MB total)
- TTS voice avatars: 8 WebP images (180KB total)
- Browser image compression lib: 57KB
- Policy pages: agreement, claw disclaimer, cookie policy
- Preview HTML: preview-1.0.0.html
- SVG icons: loadFailed.svg

### From alsgp0.fds.api.xiaomi.com (4 files)
- Scene reference images (Math, English, Travel, Beach)

### From ssl-cdn.static.browser.mi-img.com (1 file)
- pubsub.js: Xiaomi analytics/tracking

---

## Key Technical Details

### SSE Format (Legacy HTTP Chat)
```
id:msgId
event:dialogId
data:{"conversationId":"..."}

id:msgId
event:doc
data:{"content":"Hello"}

id:msgId
event:web_search
data:{"query":"...", "webSearchResults":[...]}

id:msgId
event:usage
data:{"promptTokens":100, "completionTokens":50}

id:msgId
event:finish
data:{"conversationId":"...", "msgId":"..."}
```

### WebSocket RPC Format
```json
// Send message
{
  "method": "chat.send",
  "data": {
    "sessionKey": "abc123",
    "message": "Hello",
    "deliver": false,
    "idempotencyKey": "uuid-1234"
  }
}

// Receive stream event
{
  "stream": "chat",
  "data": {
    "state": "delta",
    "content": "Hello",
    "runId": "run-123"
  }
}

// Tool call event
{
  "stream": "tool",
  "data": {
    "name": "web_search",
    "args": {"query": "weather"},
    "output": "Sunny, 25°C",
    "phase": "result"
  }
}
```

---

## Recommendations

### For Programmatic Access
1. **Use WebSocket** for chat (not HTTP REST) — the WS proxy is the primary chat interface
2. **Get Claw provisioned** for tool calling — requires gated access
3. **Use OpenRouter** (`xiaomi/mimo-v2.5-pro`) for HTTP-compatible tool calling
4. **Monitor `/ws/proxy`** events for real-time streaming responses

### For Tool Calling
1. **Option A**: Apply for Claw agent access at `aistudio.xiaomimimo.com` → Claw tab
2. **Option B**: Use OpenRouter API with MiMo model (supports `tools` parameter)
3. **Option C**: Self-host MiMo model via HuggingFace/vLLM (full control)

### For TTS
1. Use the HTTP API: `POST /open-apis/tts/v2/generate` → poll → download
2. 8 built-in voices available
3. Custom voice cloning supported via upload

---

## Files Delivered

```
/home/chaim/claw-code-go/mimo-cli/output/
├── REPORT.md                          # This report
├── api-map.json                       # Full reverse-engineering data
└── assets/
    ├── js/                            # 14 JS bundles (6.1MB)
    ├── css/                           # 4 CSS files (2.2MB)
    ├── cdn/                           # 27 CDN assets (TTS, policies, FFmpeg)
    ├── xiaomi-cdn/                     # 4 scene images
    └── misc/                          # pubsub.js
```

---

*Generated by OpenCode reverse engineering session, 2026-06-01*
