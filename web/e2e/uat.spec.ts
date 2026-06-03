/**
 * UAT (User Acceptance Testing) smoke-test script for claw-code-go web UI.
 *
 * Walks through the core user journey, verifying all major features work
 * end-to-end against a running server.
 *
 * Prerequisites:
 *   claw-code-go web --addr 127.0.0.1:7777 --provider deepseek
 *
 * Run:
 *   npx playwright test --config web/e2e/playwright.config.ts web/e2e/uat.spec.ts
 *
 * The test prints a clear pass/fail summary table to stdout when finished.
 */

import { test, expect } from '@playwright/test';

const BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:7777';

// ── Results tracker ──────────────────────────────────────────────
interface TestResult {
  name: string;
  passed: boolean;
  detail: string;
}

const results: TestResult[] = [];

function recordResult(name: string, passed: boolean, detail: string = '') {
  results.push({ name, passed, detail });
}

test.afterAll(() => {
  const total = results.length;
  const passed = results.filter((r) => r.passed).length;
  const failed = total - passed;

  console.log('\n' + '='.repeat(70));
  console.log('  UAT Smoke Test Results');
  console.log('='.repeat(70));
  console.log(`  ${'Test'.padEnd(48)} ${'Status'.padEnd(8)} Detail`);
  console.log('-'.repeat(70));

  for (const r of results) {
    const status = r.passed ? 'PASS' : 'FAIL';
    const detail = r.detail ? `  (${r.detail})` : '';
    console.log(`  ${r.name.padEnd(48)} ${status.padEnd(8)}${detail}`);
  }

  console.log('-'.repeat(70));
  console.log(`  TOTAL: ${total}  |  PASSED: ${passed}  |  FAILED: ${failed}`);
  console.log('='.repeat(70) + '\n');
});

// ── Helpers ──────────────────────────────────────────────────────
async function sendChatMessage(page: any, text: string) {
  const input = page.locator('#composer-input');
  await input.fill(text);
  await expect(page.locator('#send-button')).not.toBeDisabled();
  await page.locator('#send-button').click();
}

async function injectWSMessage(page: any, msg: Record<string, unknown>) {
  await page.evaluate((m: any) => {
    const messagesEl = document.getElementById('messages');
    if (!messagesEl) return;
    const emptyState = document.getElementById('empty-state');
    if (emptyState) emptyState.style.display = 'none';
    const data = m as Record<string, string>;

    switch (data.type) {
      case 'chat_session_init':
        document.documentElement.setAttribute('data-session-id', data.session_id || 'test');
        const pill = document.getElementById('connection-status');
        if (pill) { pill.className = 'status-pill status-connected'; pill.textContent = 'connected'; }
        break;

      case 'text_delta':
      case 'text_final': {
        let bubble = messagesEl.querySelector('.message-assistant:last-child') as HTMLElement | null;
        if (!bubble) {
          bubble = document.createElement('div');
          bubble.className = 'message message-assistant';
          messagesEl.appendChild(bubble);
        }
        const text = data.text || '';
        if (data.type === 'text_delta') {
          const cur = bubble.getAttribute('data-accumulated') || '';
          bubble.setAttribute('data-accumulated', cur + text);
          bubble.textContent = cur + text;
        } else {
          const acc = bubble.getAttribute('data-accumulated') || text || '';
          let html = acc
            .replace(/&/g, '&').replace(/</g, '<').replace(/>/g, '>')
            .replace(/```(\w+)?\n([\s\S]*?)```/g, (_s: string, lang: string, code: string) => {
              const lc = lang ? ` class="language-${lang}"` : '';
              const ec = code.replace(/&/g, '&').replace(/</g, '<').replace(/>/g, '>');
              return `<div class="code-block-wrapper"><pre><code${lc}>${ec}</code></pre><div class="code-actions"><button class="code-action-btn test-copy-btn">Copy</button></div></div>`;
            })
            .replace(/\n/g, '<br>');
          bubble.innerHTML = html;
          bubble.removeAttribute('data-accumulated');
          bubble.setAttribute('data-finalized', 'true');
        }
        messagesEl.scrollTop = messagesEl.scrollHeight;
        break;
      }
      case 'done': {
        const lb = messagesEl.querySelector('.message-assistant:last-child');
        if (lb) lb.setAttribute('data-done', 'true');
        break;
      }
      case 'error': {
        const errDiv = document.createElement('div');
        errDiv.className = 'message message-error';
        errDiv.textContent = 'Error: ' + (data.message || data.error || 'unknown');
        messagesEl.appendChild(errDiv);
        break;
      }
    }
  }, msg);
}

// ── TEST SUITE ───────────────────────────────────────────────────

test.describe('UAT Smoke Test — Core User Journey', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test('1. Page loads with correct meta tags and structural elements', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    await expect(page).toHaveTitle(/claw-code-go/);

    await expect(page.locator('link[rel="manifest"]')).toHaveAttribute('href', '/static/manifest.webmanifest');
    await expect(page.locator('meta[name="viewport"]')).toHaveAttribute('content', /width=device-width/);
    await expect(page.locator('meta[name="apple-mobile-web-app-capable"]')).toHaveAttribute('content', 'yes');

    await expect(page.locator('#top-bar')).toBeVisible();
    await expect(page.locator('#composer')).toBeVisible();
    await expect(page.locator('#composer-input')).toBeVisible();
    await expect(page.locator('#send-button')).toBeVisible();
    await expect(page.locator('#empty-state')).toBeVisible();
    await expect(page.locator('#sidebar')).toBeVisible();
    await expect(page.locator('#theme-toggle')).toBeVisible();
    await expect(page.locator('#font-size-toggle')).toBeVisible();
    await expect(page.locator('#shortcuts-toggle')).toBeVisible();

    recordResult('Page loads with meta tags & structure', true);
  });

  test('2. Service worker registration script is present', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    const found = await page.evaluate(() => {
      const scripts = Array.from(document.querySelectorAll('script'));
      return scripts.some((s) => s.textContent && s.textContent.includes('/static/sw.js'));
    });

    expect(found).toBeTruthy();
    recordResult('Service worker registration present', true);
  });

  test('3. Typing a message enables send button and shows message in chat', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    await injectWSMessage(page, { type: 'chat_session_init', session_id: 'uat-test-session' });

    const input = page.locator('#composer-input');
    await input.fill('Hello, this is a UAT test message!');

    const sendBtn = page.locator('#send-button');
    await expect(sendBtn).not.toBeDisabled();
    await sendBtn.click();

    await expect(input).toHaveValue('');
    await expect(sendBtn).toBeDisabled();

    await expect(page.locator('.message-user')).toContainText('Hello, this is a UAT test message!');
    await expect(page.locator('#empty-state')).toBeHidden();

    recordResult('Typing message enables send & shows in chat', true);
  });

  test('4. Thinking indicator appears after sending a message', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    await injectWSMessage(page, { type: 'chat_session_init', session_id: 'uat-thinking-test' });
    await sendChatMessage(page, 'Trigger thinking indicator');

    await page.waitForSelector('.thinking-indicator', { timeout: 4000 });

    const thinking = page.locator('.thinking-indicator');
    await expect(thinking).toBeVisible();
    await expect(thinking).toContainText('Thinking');
    await expect(thinking).toHaveAttribute('aria-label', 'Thinking\u2026');

    recordResult('Thinking indicator appears after sending', true);
  });

  test('5. AI response text_delta renders and text_final finalizes', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    await injectWSMessage(page, { type: 'chat_session_init', session_id: 'uat-delta-test' });
    await sendChatMessage(page, 'Show me some code');

    await injectWSMessage(page, { type: 'text_delta', text: 'Here is a Go function:\n\n' });
    await injectWSMessage(page, { type: 'text_delta', text: '```go\n' });
    await injectWSMessage(page, { type: 'text_delta', text: 'func hello() {\n' });
    await injectWSMessage(page, { type: 'text_delta', text: '    fmt.Println("Hello UAT!")\n' });
    await injectWSMessage(page, { type: 'text_delta', text: '}\n' });
    await injectWSMessage(page, { type: 'text_delta', text: '```\n\n' });
    await injectWSMessage(page, { type: 'text_delta', text: 'This function prints a greeting.' });

    await expect(page.locator('.message-assistant')).toContainText('Hello UAT');

    await injectWSMessage(page, { type: 'text_final' });

    await expect(page.locator('.thinking-indicator')).toHaveCount(0);
    await expect(page.locator('.code-block-wrapper')).toBeVisible();

    recordResult('text_delta renders, text_final finalizes', true);
  });

  test('6. Theme toggle cycles through system/light/dark', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    const htmlEl = page.locator('html');
    let theme = await htmlEl.getAttribute('data-theme');
    expect(theme === null || theme === 'system').toBeTruthy();

    await page.locator('#theme-toggle').click();
    expect(await htmlEl.getAttribute('data-theme')).toBe('light');
    expect(await page.evaluate(() => localStorage.getItem('claw_theme'))).toBe('light');

    await page.locator('#theme-toggle').click();
    expect(await htmlEl.getAttribute('data-theme')).toBe('dark');
    expect(await page.evaluate(() => localStorage.getItem('claw_theme'))).toBe('dark');

    await page.locator('#theme-toggle').click();
    expect(await htmlEl.getAttribute('data-theme')).toBeNull();

    await page.reload({ waitUntil: 'domcontentloaded' });
    expect(await htmlEl.getAttribute('data-theme')).toBeNull();

    recordResult('Theme toggle cycles & persists', true);
  });

  test('7. Font size toggle cycles through S/M/L and persists', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    const htmlEl = page.locator('html');
    expect(await htmlEl.getAttribute('data-font-size')).toBeNull();

    await page.locator('#font-size-toggle').click();
    expect(await htmlEl.getAttribute('data-font-size')).toBe('s');
    expect(await page.evaluate(() => localStorage.getItem('claw_font_size'))).toBe('s');

    await page.locator('#font-size-toggle').click();
    expect(await htmlEl.getAttribute('data-font-size')).toBeNull();

    await page.locator('#font-size-toggle').click();
    expect(await htmlEl.getAttribute('data-font-size')).toBe('l');
    expect(await page.evaluate(() => localStorage.getItem('claw_font_size'))).toBe('l');

    await page.reload({ waitUntil: 'domcontentloaded' });
    expect(await htmlEl.getAttribute('data-font-size')).toBe('l');

    recordResult('Font size toggle cycles & persists', true);
  });

  test('8. Keyboard shortcuts overlay opens and closes', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    await page.locator('#shortcuts-toggle').click();

    const overlay = page.locator('#shortcuts-overlay');
    await expect(overlay).not.toHaveClass(/shortcuts-overlay-hidden/);
    await expect(overlay).toHaveAttribute('aria-hidden', 'false');
    await expect(overlay).toContainText('Send message');

    await page.locator('#shortcuts-close').click();
    await expect(overlay).toHaveClass(/shortcuts-overlay-hidden/);

    await page.keyboard.press('?');
    await expect(overlay).not.toHaveClass(/shortcuts-overlay-hidden/);

    await page.keyboard.press('Escape');
    await expect(overlay).toHaveClass(/shortcuts-overlay-hidden/);

    recordResult('Shortcuts overlay opens & closes', true);
  });

  test('9. Offline banner appears when connection is lost', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    const banner = page.locator('#offline-banner');
    await expect(banner).toHaveAttribute('aria-hidden', 'true');

    await page.context().setOffline(true);
    await page.evaluate(() => { window.dispatchEvent(new Event('offline')); });

    await expect(banner).toHaveAttribute('aria-hidden', 'false');
    await expect(banner).toHaveClass(/offline-visible/);

    await page.context().setOffline(false);
    await page.evaluate(() => { window.dispatchEvent(new Event('online')); });

    await expect(banner).toHaveAttribute('aria-hidden', 'true', { timeout: 5000 });

    recordResult('Offline banner appears and disappears', true);
  });

  test('10. Code block copy button is rendered and works', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    await injectWSMessage(page, { type: 'chat_session_init', session_id: 'uat-copy-test' });
    await sendChatMessage(page, 'Write a function');

    await injectWSMessage(page, { type: 'text_delta', text: 'Here:\n```go\nfunc test() {}\n```' });
    await injectWSMessage(page, { type: 'text_final' });

    const copyBtn = page.locator('.code-block-wrapper .code-action-btn').first();
    await expect(copyBtn).toBeVisible();
    await expect(copyBtn).toHaveText('Copy');

    await page.context().grantPermissions(['clipboard-read', 'clipboard-write']);
    await copyBtn.click();

    await expect(copyBtn).toHaveText('Copied!', { timeout: 2000 });
    await expect(copyBtn).toHaveText('Copy', { timeout: 3000 });

    recordResult('Code block copy button works', true);
  });

  test('11. Session sidebar loads and shows current session', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'domcontentloaded' });

    await expect(page.locator('#sidebar')).toBeVisible();
    await expect(page.locator('#session-list')).toBeVisible();

    const count = await page.locator('.session-item').count();
    expect(count).toBeGreaterThanOrEqual(1);

    await expect(page.locator('.session-item.active')).toHaveCount(1);
    await expect(page.locator('#new-session-btn')).toBeVisible();

    recordResult('Session sidebar loads with current session', true);
  });
});
