// Puppeteer browser setup with auth cookie injection and stealth
import puppeteer from 'puppeteer';

const DEFAULT_VIEWPORT = { width: 1440, height: 900 };

const STEALTH_SCRIPT = `
  // Pass basic CDP-driven detection tests
  Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
  window.chrome = { runtime: {}, loadTimes: () => ({}), csi: () => ({}) };
  const origQuery = window.navigator.permissions?.query;
  if (origQuery) {
    window.navigator.permissions.query = (p) =>
      p.name === 'notifications' ? Promise.resolve({ state: Notification.permission }) : origQuery(p);
  }
  const getParameter = WebGLRenderingContext.prototype.getParameter;
  WebGLRenderingContext.prototype.getParameter = function(p) {
    if (p === 37445) return 'Intel Inc.';
    if (p === 37446) return 'Intel Iris OpenGL Engine';
    return getParameter.call(this, p);
  };
`;

export async function launchBrowser({ headless = true, stealth = true } = {}) {
  const browser = await puppeteer.launch({
    headless: headless ? 'new' : false,
    args: [
      '--no-sandbox',
      '--disable-setuid-sandbox',
      '--disable-blink-features=AutomationControlled',
      '--disable-dev-shm-usage',
      '--window-size=1440,900',
      '--lang=en-US,en',
    ],
    defaultViewport: DEFAULT_VIEWPORT,
  });
  return browser;
}

export async function newPage(browser, { stealth = true, cookies, baseUrl, verbose = false } = {}) {
  const page = await browser.newPage();
  if (stealth) {
    await page.evaluateOnNewDocument(STEALTH_SCRIPT);
  }
  // Set realistic UA and locale
  await page.setUserAgent(
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36'
  );
  await page.setExtraHTTPHeaders({
    'accept-language': 'en-US,en;q=0.9',
    'x-timezone': 'Asia/Jerusalem',
  });

  // Inject auth cookies on the right domain
  if (cookies && baseUrl) {
    const u = new URL(baseUrl);
    const cookieList = [
      { name: 'serviceToken', value: cookies.serviceToken, domain: u.hostname, path: '/', httpOnly: false, secure: true, sameSite: 'None' },
      { name: 'userId', value: cookies.userId, domain: u.hostname, path: '/', httpOnly: false, secure: true, sameSite: 'None' },
      { name: 'xiaomichatbot_ph', value: cookies.xiaomichatbot_ph, domain: u.hostname, path: '/', httpOnly: false, secure: true, sameSite: 'None' },
    ];
    await browser.setCookie(...cookieList);
    if (verbose) console.log(`[browser] injected ${cookieList.length} cookies for ${u.hostname}`);
  }

  return page;
}
