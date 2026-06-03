/**
 * Responsive design visual regression tests.
 *
 * Captures screenshots of the chat UI at 5 breakpoints (mobile → desktop)
 * in both light and dark mode. Stores baselines in web/e2e/baselines/responsive/.
 *
 * Run:
 *   npx playwright test --config web/e2e/playwright.config.ts web/e2e/responsive.spec.ts
 */

import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:7777';
const RESPONSIVE_DIR = path.join(__dirname, 'baselines', 'responsive');

if (!fs.existsSync(RESPONSIVE_DIR)) {
  fs.mkdirSync(RESPONSIVE_DIR, { recursive: true });
}

// Five breakpoints: mobile, iPhone 14, tablet portrait, tablet landscape, desktop
const BREAKPOINTS = [
  { name: 'mobile-375', width: 375, height: 812 },
  { name: 'iphone14-390', width: 390, height: 844 },
  { name: 'tablet-portrait-768', width: 768, height: 1024 },
  { name: 'tablet-landscape-1024', width: 1024, height: 768 },
  { name: 'desktop-1440', width: 1440, height: 900 },
] as const;

const THEMES = ['light', 'dark'] as const;

async function expectScreenshot(
  page: import('@playwright/test').Page,
  name: string,
): Promise<void> {
  const screenshot = await page.screenshot({ fullPage: true });
  const baselinePath = path.join(RESPONSIVE_DIR, `${name}.png`);

  if (!fs.existsSync(baselinePath)) {
    fs.writeFileSync(baselinePath, screenshot);
    test.fail(
      true,
      `No baseline for "${name}". New baseline saved to ${baselinePath}. Review and re-run.`,
    );
    return;
  }

  expect(screenshot).toMatchSnapshot({
    name: `responsive/${name}`,
    maxDiffPixelRatio: 0.02,
  });
}

for (const bp of BREAKPOINTS) {
  for (const theme of THEMES) {
    test.describe(`Responsive — ${bp.name} / ${theme}`, () => {
      test.use({
        viewport: { width: bp.width, height: bp.height },
        colorScheme: theme === 'dark' ? 'dark' : 'light',
      });

      test('empty state', async ({ page }) => {
        await page.goto(BASE_URL, { waitUntil: 'networkidle' });

        await expect(page.locator('#empty-state')).toBeVisible();
        await expect(page.locator('#top-bar')).toBeVisible();
        await expect(page.locator('#composer')).toBeVisible();
        await expect(page.locator('#send-button')).toBeVisible();

        await expectScreenshot(page, `${bp.name}-${theme}-empty`);
      });

      test('composer interaction', async ({ page }) => {
        await page.goto(BASE_URL, { waitUntil: 'networkidle' });

        const input = page.locator('#composer-input');
        await input.fill('Testing responsive layout');

        const sendBtn = page.locator('#send-button');
        await expect(sendBtn).not.toBeDisabled();

        await expectScreenshot(page, `${bp.name}-${theme}-composer`);
      });

      test('sidebar toggle', async ({ page }) => {
        await page.goto(BASE_URL, { waitUntil: 'networkidle' });

        // On mobile, sidebar is hidden by default
        const sidebar = page.locator('#sidebar');
        const toggle = page.locator('#sidebar-toggle');

        if (bp.width < 768) {
          await expect(sidebar).not.toHaveClass(/sidebar-open/);
          await toggle.click();
          await expect(sidebar).toHaveClass(/sidebar-open/);
          await expectScreenshot(page, `${bp.name}-${theme}-sidebar-open`);
          await toggle.click();
        } else {
          await expect(sidebar).toBeVisible();
        }
      });

      test('shortcuts overlay', async ({ page }) => {
        await page.goto(BASE_URL, { waitUntil: 'networkidle' });

        await page.locator('#shortcuts-toggle').click();
        const overlay = page.locator('#shortcuts-overlay');
        await expect(overlay).not.toHaveClass(/shortcuts-overlay-hidden/);

        await expectScreenshot(page, `${bp.name}-${theme}-shortcuts`);

        await page.keyboard.press('Escape');
        await expect(overlay).toHaveClass(/shortcuts-overlay-hidden/);
      });

      // Only run conversation test on larger viewports to avoid slow API calls
      if (bp.width >= 768) {
        test('conversation with messages', async ({ page }) => {
          await page.goto(BASE_URL, { waitUntil: 'networkidle' });

          // Inject messages to see layout with content
          await page.evaluate(() => {
            const el = document.getElementById('messages');
            if (!el) return;
            const empty = document.getElementById('empty-state');
            if (empty) empty.style.display = 'none';

            for (let i = 0; i < 3; i++) {
              const userMsg = document.createElement('div');
              userMsg.className = 'message message-user';
              userMsg.textContent = `User message ${i + 1}`;
              el.appendChild(userMsg);

              const aiMsg = document.createElement('div');
              aiMsg.className = 'message message-assistant';
              aiMsg.textContent = `AI response ${i + 1} with some longer text to demonstrate the responsive layout at this breakpoint.`;
              el.appendChild(aiMsg);
            }
          });

          await expectScreenshot(page, `${bp.name}-${theme}-conversation`);
        });
      }
    });
  }
}
