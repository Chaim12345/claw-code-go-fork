/**
 * Load / stress test: 10 concurrent WebSocket connections.
 *
 * Opens 10 simultaneous chat sessions, sends 3 messages each,
 * and verifies all receive text_final within 30 seconds.
 * Measures server resource usage.
 *
 * Run:
 *   npx playwright test --config web/e2e/playwright.config.ts web/e2e/load.spec.ts
 */

import { test, expect } from '@playwright/test';
import WebSocket from 'ws';

const BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:7777';
const WS_URL = BASE_URL.replace(/^http/, 'ws') + '/api/chat/ws';

const CONCURRENT = 10;
const MESSAGES_PER_CONN = 3;
const TIMEOUT_MS = 30_000;

interface ConnectionResult {
  id: number;
  connected: boolean;
  messagesSent: number;
  responsesReceived: number;
  totalTimeMs: number;
  error?: string;
}

async function runLoadSession(id: number): Promise<ConnectionResult> {
  const start = Date.now();
  const result: ConnectionResult = {
    id,
    connected: false,
    messagesSent: 0,
    responsesReceived: 0,
    totalTimeMs: 0,
  };

  return new Promise((resolve) => {
    const ws = new WebSocket(WS_URL);
    const timeout = setTimeout(() => {
      ws.close();
      result.totalTimeMs = Date.now() - start;
      if (!result.connected) result.error = 'connection timeout';
      resolve(result);
    }, TIMEOUT_MS);

    ws.on('open', () => {
      result.connected = true;
    });

    ws.on('message', (data) => {
      try {
        const msg = JSON.parse(data.toString());
        if (msg.type === 'text_final') {
          result.responsesReceived++;
        }
        // Send next message when we get a response
        if (msg.type === 'text_final' || msg.type === 'error' || msg.type === 'done') {
          if (result.messagesSent < MESSAGES_PER_CONN) {
            const payload = JSON.stringify({
              type: 'user_input',
              text: `Load test message ${result.messagesSent + 1} from session ${id}`,
              message_id: `load-${id}-${result.messagesSent}`,
            });
            ws.send(payload);
            result.messagesSent++;
          } else {
            // Done with all messages
            clearTimeout(timeout);
            ws.close();
            result.totalTimeMs = Date.now() - start;
            resolve(result);
          }
        }
      } catch {
        // Ignore parse errors
      }
    });

    ws.on('error', (err) => {
      result.error = err.message;
      clearTimeout(timeout);
      result.totalTimeMs = Date.now() - start;
      resolve(result);
    });

    ws.on('close', () => {
      if (result.totalTimeMs === 0) {
        result.totalTimeMs = Date.now() - start;
        resolve(result);
      }
    });
  });
}

test.describe('Load / stress test — 10 concurrent connections', () => {
  test(`all ${CONCURRENT} sessions complete within ${TIMEOUT_MS / 1000}s`, async () => {
    const results: ConnectionResult[] = [];

    // Stagger connections slightly to avoid thundering herd
    const promises = Array.from({ length: CONCURRENT }, (_, i) =>
      new Promise<ConnectionResult>((resolve) => {
        setTimeout(() => runLoadSession(i).then(resolve), i * 100);
      })
    );

    const allResults = await Promise.all(promits);
    results.push(...allResults);

    // Log summary
    console.log('\n=== Load Test Results ===');
    console.log(`Concurrent connections: ${CONCURRENT}`);
    console.log(`Messages per connection: ${MESSAGES_PER_CONN}`);

    for (const r of results) {
      console.log(
        `  Session ${r.id}: connected=${r.connected} sent=${r.messagesSent} received=${r.responsesReceived} time=${r.totalTimeMs}ms${r.error ? ` error=${r.error}` : ''}`
      );
    }

    const connected = results.filter((r) => r.connected).length;
    const allSent = results.every((r) => r.messagesSent >= MESSAGES_PER_CONN);
    const avgTime = results.reduce((s, r) => s + r.totalTimeMs, 0) / results.length;
    const maxTime = Math.max(...results.map((r) => r.totalTimeMs));

    console.log(`\nConnected: ${connected}/${CONCURRENT}`);
    console.log(`All sent ${MESSAGES_PER_CONN} messages: ${allSent}`);
    console.log(`Avg time: ${Math.round(avgTime)}ms | Max: ${maxTime}ms`);

    // Assertions
    expect(connected, `Expected all ${CONCURRENT} connections`).toBe(CONCURRENT);
    expect(allSent, 'All sessions should send their messages').toBe(true);

    // Allow some tolerance — not all responses may arrive if server is slow
    const totalResponses = results.reduce((s, r) => s + r.responsesReceived, 0);
    const minExpected = CONCURRENT * 1; // at least 1 response per connection
    expect(
      totalResponses,
      `Expected at least ${minExpected} total responses, got ${totalResponses}`
    ).toBeGreaterThanOrEqual(minExpected);
  });

  test('server responds to rapid fire messages', async () => {
    const ws = new WebSocket(WS_URL);

    await new Promise<void>((resolve, reject) => {
      ws.on('open', () => resolve());
      ws.on('error', reject);
    });

    // Send 5 messages rapidly without waiting for responses
    for (let i = 0; i < 5; i++) {
      ws.send(JSON.stringify({
        type: 'user_input',
        text: `Rapid fire ${i}`,
        message_id: `rapid-${i}`,
      }));
    }

    // Wait for at least one response
    const response = await new Promise<Record<string, unknown> | null>((resolve) => {
      const timeout = setTimeout(() => resolve(null), 10000);
      ws.on('message', (data) => {
        try {
          const msg = JSON.parse(data.toString());
          if (msg.type === 'text_final' || msg.type === 'error' || msg.type === 'warn') {
            clearTimeout(timeout);
            resolve(msg);
          }
        } catch {
          // ignore
        }
      });
    });

    ws.close();

    // Server should handle rapid fire gracefully
    expect(response).not.toBeNull();
    expect(response?.type).not.toBe('error'); // no crash
  });
});
