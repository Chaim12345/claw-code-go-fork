import { test, expect } from '@playwright/test';
import path from 'path';
import fs from 'fs';

const BASELINE_DIR = path.join(__dirname, 'baselines');
const SCREENSHOT_OPTIONS = { fullPage: true };
const MOBILE_VIEWPORT = { width: 390, height: 844 }; // iPhone 14

// Ensure baselines directory exists.
if (!fs.existsSync(BASELINE_DIR)) {
  fs.mkdirSync(BASELINE_DIR, { recursive: true });
}

// Resolve the server URL from environment or default to localhost.
const BASE_URL = process.env.E2E_BASE_URL || 'http://127.0.0.1:7777';

// Helper: compare a screenshot against a stored baseline.
// If no baseline exists, this test will fail with instructions to
// regenerate. This avoids silently passing on a missing baseline.
async function expectScreenshot(
  page: import('@playwright/test').Page,
  name: string
): Promise<void> {
  const screenshot = await page.screenshot(SCREENSHOT_OPTIONS);
  const baselinePath = path.join(BASELINE_DIR, `${name}.png`);

  if (!fs.existsSync(baselinePath)) {
    // Save the screenshot so the developer can review it.
    fs.writeFileSync(baselinePath, screenshot);
    test.fail(
      true,
      `No baseline for "${name}". A new baseline has been saved to ${baselinePath}. ` +
        `Review it and re-run the tests.`
    );
    return;
  }

  const baseline = fs.readFileSync(baselinePath);
  // Use Playwright's built-in snapshot comparison with a 2% pixel tolerance.
  expect(screenshot).toMatchSnapshot({
    name,
    maxDiffPixelRatio: 0.02, // 2% tolerance for minor rendering differences
  });
}

test.describe('Mobile visual regression — iPhone 14', () => {
  // Set the viewport for all tests in this block.
  test.use({ viewport: MOBILE_VIEWPORT });

  test('empty state renders correctly', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'networkidle' });

    // Verify the empty state is visible.
    const emptyState = page.locator('#empty-state');
    await expect(emptyState).toBeVisible();

    // Verify key structural elements are present.
    await expect(page.locator('#top-bar')).toBeVisible();
    await expect(page.locator('#composer')).toBeVisible();
    await expect(page.locator('#composer-input')).toBeVisible();
    await expect(page.locator('#send-button')).toBeVisible();

    // Verify connection status pill shows connection attempt.
    const statusPill = page.locator('#connection-status');
    await expect(statusPill).toBeVisible();

    await expectScreenshot(page, 'mobile-empty-state');
  });

  test('composer accepts text input on mobile', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'networkidle' });

    const input = page.locator('#composer-input');
    await input.fill('Hello from mobile visual regression test!');

    // Verify the send button becomes enabled.
    const sendButton = page.locator('#send-button');
    await expect(sendButton).not.toBeDisabled();

    // Verify the character count is hidden (under threshold).
    const charCount = page.locator('#char-count');
    await expect(charCount).toHaveClass(/char-count-hidden/);

    await expectScreenshot(page, 'mobile-composer-filled');
  });

  test('sidebar toggle works on mobile', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'networkidle' });

    // On mobile viewport, the sidebar should be hidden by default.
    const sidebar = page.locator('#sidebar');
    await expect(sidebar).not.toHaveClass(/sidebar-open/);

    // Tap the hamburger menu.
    const toggle = page.locator('#sidebar-toggle');
    await toggle.click();

    // Sidebar should now be open.
    await expect(sidebar).toHaveClass(/sidebar-open/);

    await expectScreenshot(page, 'mobile-sidebar-open');

    // Tap again to close.
    await toggle.click();
    await expect(sidebar).not.toHaveClass(/sidebar-open/);
  });

  test('themes render correctly in dark mode', async ({ page }) => {
    // Set dark mode via colorScheme emulation.
    await page.emulateMedia({ colorScheme: 'dark' });
    await page.goto(BASE_URL, { waitUntil: 'networkidle' });

    // In dark mode, the document background should be dark.
    const bgColor = await page.evaluate(() => {
      return getComputedStyle(document.documentElement).getPropertyValue(
        '--claw-bg'
      );
    });

    // The background should not be pure white (light mode default).
    expect(bgColor).not.toBe('#ffffff');

    await expectScreenshot(page, 'mobile-dark-mode');
  });

  test('keyboard shortcuts overlay renders on mobile', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'networkidle' });

    // Tap the "?" button in the header (touch-friendly shortcut toggle).
    const shortcutsBtn = page.locator('#shortcuts-toggle');
    await shortcutsBtn.click();

    // The overlay should be visible.
    const overlay = page.locator('#shortcuts-overlay');
    await expect(overlay).not.toHaveClass(/shortcuts-overlay-hidden/);
    await expect(overlay).toHaveAttribute('aria-hidden', 'false');

    await expectScreenshot(page, 'mobile-shortcuts-overlay');

    // Close the overlay.
    const closeBtn = page.locator('#shortcuts-close');
    await closeBtn.click();
    await expect(overlay).toHaveClass(/shortcuts-overlay-hidden/);
  });

  test('offline banner appears when disconnected', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'networkidle' });

    // Simulate going offline.
    await page.context().setOffline(true);

    // The offline banner should appear.
    const banner = page.locator('#offline-banner');
    await expect(banner).toHaveClass(/offline-visible/);
    await expect(banner).toHaveAttribute('aria-hidden', 'false');

    await expectScreenshot(page, 'mobile-offline-banner');
  });

  test('composer is sticky at bottom on mobile scroll', async ({ page }) => {
    await page.goto(BASE_URL, { waitUntil: 'networkidle' });

    // Inject enough messages to cause a scroll.
    for (let i = 0; i < 5; i++) {
      await page.evaluate((n) => {
        const el = document.getElementById('messages');
        if (!el) return;
        const div = document.createElement('div');
        div.className = 'message message-user';
        div.textContent = `Test message number ${n}`;
        el.appendChild(div);
        // Hide empty state.
        const empty = document.getElementById('empty-state');
        if (empty) empty.style.display = 'none';
      }, i);
    }

    // Scroll up a bit.
    const messages = page.locator('#messages');
    await messages.evaluate((el) => {
      el.scrollTop = 100;
    });

    // The composer (footer) should still be visible at the bottom.
    const composer = page.locator('#composer');
    await expect(composer).toBeVisible();

    await expectScreenshot(page, 'mobile-scrolled-with-messages');
  });
});