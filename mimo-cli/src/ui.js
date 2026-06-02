// UI mapper - crawl routes and capture page state
import { URL } from 'node:url';
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';

export class UiMapper {
  constructor({ base, routes = null, maxPages = 30, interact = true, verbose = false } = {}) {
    this.base = new URL(base);
    this.routes = routes;
    this.maxPages = maxPages;
    this.interact = interact;
    this.verbose = verbose;
    this.pages = [];
  }

  // Discover routes by visiting home and finding links
  async discoverRoutes(page) {
    await page.goto(this.base.href, { waitUntil: 'networkidle2', timeout: 30000 });
    await new Promise(r => setTimeout(r, 1500)); // let SPA bootstrap

    // Extract in-app links (#/...)
    const links = await page.evaluate(() => {
      const out = new Set();
      const all = Array.from(document.querySelectorAll('a, button'));
      for (const el of all) {
        const href = el.getAttribute('href') || '';
        if (href.startsWith('#/') || href.startsWith('/#')) {
          const route = href.replace(/^#?\/?#?\/?/, '/');
          out.add(route);
        }
        // Buttons with onclick navigation
        const onClick = el.getAttribute('onclick') || '';
        const m = onClick.match(/['"](\/[a-zA-Z0-9\-_/]+)['"]/);
        if (m) out.add(m[1]);
      }
      return [...out];
    });
    const initial = ['/', '/c', ...links];
    const unique = [...new Set(initial)].filter(r => r.startsWith('/')).slice(0, this.maxPages);
    if (this.verbose) console.log(`[ui] discovered ${unique.length} routes`);
    return unique;
  }

  async visitRoute(page, route, pageId) {
    const target = new URL(route, this.base).href;
    const meta = { id: pageId, route, url: target, ts: Date.now(), snapshot: null, api: [], assets: [], errors: [] };
    try {
      const resp = await page.goto(target, { waitUntil: 'networkidle2', timeout: 30000 });
      meta.status = resp?.status();
      meta.title = await page.title();
      await new Promise(r => setTimeout(r, 1500));
      meta.snapshot = await this.snapshot(page);
    } catch (e) {
      meta.error = e.message;
    }
    return meta;
  }

  async snapshot(page) {
    return page.evaluate(() => {
      const out = {
        title: document.title,
        url: location.href,
        h1: Array.from(document.querySelectorAll('h1,h2')).slice(0, 8).map(h => h.innerText.trim()).filter(Boolean),
        bodyText: document.body.innerText.substring(0, 2000),
        interactive: [],
        forms: [],
        sections: [],
      };
      // buttons + links
      out.interactive = Array.from(document.querySelectorAll('button, a[href]'))
        .slice(0, 60)
        .map(el => {
          const rect = el.getBoundingClientRect();
          return {
            tag: el.tagName,
            text: (el.innerText || el.textContent || '').trim().substring(0, 80),
            href: el.getAttribute('href') || null,
            aria: el.getAttribute('aria-label') || null,
            cls: (el.className || '').toString().substring(0, 80),
            visible: rect.width > 0 && rect.height > 0,
          };
        })
        .filter(el => el.text || el.href);
      out.forms = Array.from(document.querySelectorAll('form,textarea,input,select')).map(el => ({
        tag: el.tagName,
        type: el.type || null,
        name: el.name || null,
        placeholder: el.placeholder || null,
        aria: el.getAttribute('aria-label') || null,
      }));
      // Section headings
      out.sections = Array.from(document.querySelectorAll('h1,h2,h3,h4,[role="heading"]'))
        .slice(0, 30)
        .map(h => ({ level: h.tagName, text: h.innerText.trim().substring(0, 120) }))
        .filter(h => h.text);
      // Count components by class prefix
      const counters = {};
      for (const el of document.querySelectorAll('[class*="-"]')) {
        const cls = (el.className || '').toString();
        for (const m of cls.matchAll(/(arco-[a-z\-]+|ant-[a-z\-]+|mimo-[a-z\-]+)/g)) {
          counters[m[1]] = (counters[m[1]] || 0) + 1;
        }
      }
      out.componentCounts = Object.entries(counters).sort((a, b) => b[1] - a[1]).slice(0, 20);
      return out;
    });
  }

  async interactWithPage(page) {
    if (!this.interact) return;
    try {
      // Try clicking the chat send button area
      const ta = await page.$('textarea');
      if (ta) {
        await ta.focus();
        await page.keyboard.type('hello');
        await new Promise(r => setTimeout(r, 200));
        // don't actually send - we don't want to clutter the conversation
      }
    } catch (e) {
      if (this.verbose) console.log(`[ui] interact error: ${e.message}`);
    }
  }

  async crawl(page, recorder) {
    let routes = this.routes;
    if (!routes) {
      routes = await this.discoverRoutes(page);
    } else {
      // Just visit home first so cookies are set
      await page.goto(this.base.href, { waitUntil: 'networkidle2', timeout: 30000 });
      await new Promise(r => setTimeout(r, 1500));
    }

    for (let i = 0; i < routes.length; i++) {
      const route = routes[i];
      if (this.verbose) console.log(`[ui] visiting [${i + 1}/${routes.length}] ${route}`);
      const pageMeta = { id: i + 1, route };
      recorder.setPage(pageMeta);
      const visited = await this.visitRoute(page, route, i + 1);
      this.pages.push(visited);
      // After the page is visited, the recorder has captured the network for it
      visited.api = recorder.apiCalls.filter(c => c.page === i + 1).map(c => ({ url: c.url, method: c.method, status: c.status, hasBody: !!c.postData }));
      visited.assets = recorder.assets.filter(a => a.page === i + 1).map(a => ({ url: a.url, type: a.resourceType }));
      await this.interactWithPage(page);
    }
    return this.pages;
  }
}
