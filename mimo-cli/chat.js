#!/usr/bin/env node
/**
 * MiMo Studio CLI Chat
 * 
 * Interactive CLI for chatting with MiMo models via HTTP SSE.
 * 
 * Usage:
 *   node chat.js                    # Default: mimo-v2.5-pro
 *   node chat.js --model mimo-v2-pro
 *   node chat.js --thinking         # Enable deep thinking
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import readline from 'readline';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// ── Config ──────────────────────────────────────────────────────────────────

const HOST = 'https://aistudio.xiaomimimo.com';
const AUTH_PATH = path.join(__dirname, 'auth.json');

const DEFAULTS = {
  model: 'mimo-v2.5-pro',
  thinking: false,
  webSearchMode: 'disabled',
};

// ── Parse args ──────────────────────────────────────────────────────────────

function parseArgs() {
  const args = process.argv.slice(2);
  const opts = { ...DEFAULTS };
  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--model':
      case '-m':
        opts.model = args[++i];
        break;
      case '--thinking':
      case '-t':
        opts.thinking = true;
        break;
      case '--session':
      case '-s':
        opts.sessionKey = args[++i];
        break;
      case '--search':
        opts.webSearchMode = 'auto';
        break;
      case '--help':
      case '-h':
        console.log(`
MiMo Studio CLI Chat

Usage:
  node chat.js [options]

Options:
  -m, --model <id>      Model ID (default: mimo-v2.5-pro)
                        Available: mimo-v2-flash, mimo-v2-flash-studio,
                                   mimo-v2-pro, mimo-v2.5, mimo-v2.5-pro
  -t, --thinking        Enable deep thinking / reasoning
  -s, --session <key>   Resume a specific session
  --search              Enable web search
  -h, --help            Show this help

Commands (while chatting):
  /model <id>           Switch model
  /thinking             Toggle thinking
  /search               Toggle web search
  /session              Show current session key
  /new                  Start new session
  /clear                Clear screen
  /quit                 Exit
`);
        process.exit(0);
    }
  }
  return opts;
}

// ── Auth ────────────────────────────────────────────────────────────────────

function loadAuth() {
  if (!fs.existsSync(AUTH_PATH)) {
    console.error(`Error: Auth file not found at ${AUTH_PATH}`);
    process.exit(1);
  }
  return JSON.parse(fs.readFileSync(AUTH_PATH, 'utf-8'));
}

function getCookieString(auth) {
  return [
    `serviceToken=${auth.serviceToken}`,
    `userId=${auth.userId}`,
    `xiaomichatbot_ph=${auth.xiaomichatbot_ph}`,
  ].join('; ');
}

// ── Chat via HTTP SSE ───────────────────────────────────────────────────────

async function chat(auth, opts, message, conversationId) {
  const msgId = crypto.randomUUID();
  const ph = auth.xiaomichatbot_ph;

  const body = {
    model: opts.model,
    msgId,
    message,
    webSearchMode: opts.webSearchMode,
    enableThinking: opts.thinking,
    temperature: 0.7,
    topP: 0.9,
  };
  if (conversationId) body.conversationId = conversationId;

  const res = await fetch(`${HOST}/open-apis/bot/chat?xiaomichatbot_ph=${encodeURIComponent(ph)}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Cookie: getCookieString(auth),
      'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36',
      Referer: `${HOST}/`,
    },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Chat request failed (${res.status}): ${text}`);
  }

  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = '';
  let result = { conversationId, content: '', webSearchResults: [], usage: null };

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    buffer += decoder.decode(value, { stream: true });
    const lines = buffer.split('\n');
    buffer = lines.pop(); // Keep incomplete line

    let currentEvent = null;
    let currentData = '';

    for (const line of lines) {
      if (line.startsWith('event:')) {
        currentEvent = line.slice(6).trim();
      } else if (line.startsWith('data:')) {
        currentData = line.slice(5).trim();
      } else if (line === '' && currentEvent && currentData) {
        // Process event
        try {
          const data = JSON.parse(currentData);
          switch (currentEvent) {
            case 'dialogId':
              result.conversationId = data.conversationId;
              break;
            case 'doc':
              if (data.content) {
                process.stdout.write(data.content);
                result.content += data.content;
              }
              break;
            case 'web_search':
              if (data.webSearchResults) {
                result.webSearchResults = data.webSearchResults;
              }
              break;
            case 'usage':
              result.usage = data;
              break;
            case 'finish':
              // Done
              break;
          }
        } catch (e) {
          // Skip malformed JSON
        }
        currentEvent = null;
        currentData = '';
      }
    }
  }

  return result;
}

// ── CLI REPL ────────────────────────────────────────────────────────────────

async function main() {
  const opts = parseArgs();
  const auth = loadAuth();

  console.log('========================================');
  console.log('       MiMo Studio CLI Chat');
  console.log('========================================');
  console.log();
  console.log(`  Model:    ${opts.model}`);
  console.log(`  Thinking: ${opts.thinking ? 'ON' : 'off'}`);
  console.log(`  Search:   ${opts.webSearchMode}`);
  console.log();
  console.log('  Type your message, or /help for commands.');
  console.log('  Ctrl+C to exit.');
  console.log();

  let conversationId = null;

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    prompt: '\nYou> ',
  });

  function showPrompt() {
    rl.setPrompt('\nYou> ');
    rl.prompt();
  }

  rl.on('line', async (line) => {
    const trimmed = line.trim();

    if (trimmed.startsWith('/')) {
      const [cmd, ...args] = trimmed.split(/\s+/);

      switch (cmd.toLowerCase()) {
        case '/help':
          console.log(`
  Commands:
    /model <id>     Switch model (e.g., /model mimo-v2-pro)
    /thinking       Toggle deep thinking
    /search         Toggle web search
    /session        Show session ID
    /new            Start new session
    /clear          Clear screen
    /quit           Exit
`);
          showPrompt();
          return;

        case '/model':
        case '/m':
          if (args[0]) {
            opts.model = args[0];
            console.log(`  Model: ${opts.model}`);
          } else {
            console.log(`  Model: ${opts.model}`);
          }
          showPrompt();
          return;

        case '/thinking':
        case '/t':
          opts.thinking = !opts.thinking;
          console.log(`  Thinking: ${opts.thinking ? 'ON' : 'off'}`);
          showPrompt();
          return;

        case '/search':
          opts.webSearchMode = opts.webSearchMode === 'disabled' ? 'auto' : 'disabled';
          console.log(`  Web search: ${opts.webSearchMode}`);
          showPrompt();
          return;

        case '/session':
          console.log(`  Session: ${conversationId || '(new)'}`);
          showPrompt();
          return;

        case '/new':
          conversationId = null;
          console.log('  New session started.');
          showPrompt();
          return;

        case '/clear':
          console.clear();
          showPrompt();
          return;

        case '/quit':
        case '/exit':
          console.log('  Bye!');
          process.exit(0);

        default:
          console.log(`  Unknown command: ${cmd} (type /help)`);
          showPrompt();
          return;
      }
    }

    if (!trimmed) {
      showPrompt();
      return;
    }

    process.stdout.write('\nMiMo> ');
    try {
      const result = await chat(auth, opts, trimmed, conversationId);
      if (result.conversationId && !conversationId) {
        conversationId = result.conversationId;
      }
      if (result.webSearchResults?.length > 0) {
        console.log(`\n  [${result.webSearchResults.length} sources found]`);
      }
      if (result.usage) {
        const u = result.usage;
        console.log(`\n  [tokens: ${u.promptTokens || 0} in / ${u.completionTokens || 0} out]`);
      }
    } catch (err) {
      console.error(`\n  Error: ${err.message}`);
    }
    console.log();
    showPrompt();
  });

  rl.on('close', () => {
    console.log('\n  Bye!');
    process.exit(0);
  });

  process.on('SIGINT', () => {
    console.log('\n  Bye!');
    process.exit(0);
  });

  showPrompt();
}

main().catch((err) => {
  console.error(`Fatal: ${err.message}`);
  process.exit(1);
});
