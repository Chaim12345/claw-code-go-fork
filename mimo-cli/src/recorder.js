// Network interceptor - captures every request/response/SSE event
import { URL } from 'node:url';

const STATIC_EXT = /\.(js|css|ts|tsx|jsx|map|png|jpg|jpeg|webp|gif|svg|ico|woff2?|ttf|otf|eot|mp4|mp3|webm|wasm)(\?|$)/i;
const API_PATTERN = /^https?:\/\/[^/]+\/open-apis\//i;

export class ApiRecorder {
  constructor({ maxBodyBytes = 64 * 1024, maxEvents = 5000, verbose = false } = {}) {
    this.maxBodyBytes = maxBodyBytes;
    this.maxEvents = maxEvents;
    this.verbose = verbose;
    this.apiCalls = [];
    this.assets = [];
    this.pages = [];
    this.currentPage = null;
    this._seq = 0;
  }

  attach(page) {
    page.on('request', (req) => this._onRequest(req));
    page.on('requestfinished', (req) => this._onRequestFinished(req));
    page.on('requestfailed', (req) => this._onRequestFailed(req));
    page.on('response', (res) => this._onResponse(res));
    page.on('console', (msg) => this._onConsole(msg));
    page.on('pageerror', (err) => this._onPageError(err));
    page.on('framenavigated', (frame) => this._onFrameNavigated(frame));
  }

  setPage(pageMeta) {
    this.currentPage = pageMeta;
  }

  _onRequest(req) {
    const url = req.url();
    const method = req.method();
    const isApi = API_PATTERN.test(url);
    const isStatic = STATIC_EXT.test(url);
    const resourceType = req.resourceType();

    const entry = {
      id: ++this._seq,
      ts: Date.now(),
      page: this.currentPage?.id || null,
      pageUrl: this.currentPage?.url || null,
      method,
      url,
      resourceType,
      isApi,
      isStatic,
      headers: req.headers(),
      postData: null,
    };
    try {
      const pd = req.postData();
      if (pd) entry.postData = pd.length > this.maxBodyBytes ? pd.substring(0, this.maxBodyBytes) + '...[truncated]' : pd;
    } catch (e) {}

    if (isApi) {
      this.apiCalls.push(entry);
    } else if (isStatic || ['script', 'stylesheet', 'image', 'font', 'media'].includes(resourceType)) {
      this.assets.push(entry);
    }

    if (this.verbose && isApi) console.log(`[api] ${method} ${url}`);
  }

  _onRequestFinished(req) {
    const id = req._mimoId;
    const url = req.url();
    const entry = this._findEntry(url, req.method());
    if (entry) entry.finished = true;
  }

  _onRequestFailed(req) {
    const url = req.url();
    const entry = this._findEntry(url, req.method());
    if (entry) {
      entry.failed = true;
      entry.failure = req.failure()?.errorText;
    }
  }

  async _onResponse(res) {
    const url = res.url();
    const method = res.request().method();
    const entry = this._findEntry(url, method);
    if (!entry) return;

    entry.status = res.status();
    entry.statusText = res.statusText();
    entry.respHeaders = res.headers();

    const ct = res.headers()['content-type'] || '';
    entry.contentType = ct;

    // Capture body for non-streaming responses
    const isStream = /event-stream|text\/event-stream/i.test(ct);
    entry.isSSE = isStream;
    if (res.status() < 400 && !isStream && entry.isApi) {
      try {
        const buf = await res.buffer();
        const text = buf.toString('utf8');
        entry.respBody = text.length > this.maxBodyBytes ? text.substring(0, this.maxBodyBytes) + '...[truncated]' : text;
        entry.respSize = buf.length;
      } catch (e) {
        entry.respError = e.message;
      }
    } else if (isStream) {
      // For SSE, capture first chunk to see event types
      try {
        const buf = await res.buffer();
        const text = buf.toString('utf8');
        entry.respBody = text.length > 16384 ? text.substring(0, 16384) + '...[truncated]' : text;
        entry.respSize = buf.length;
        // Extract unique SSE event names
        const events = [...text.matchAll(/^event:\s*(\S+)/gm)].map(m => m[1]);
        entry.sseEvents = [...new Set(events)];
      } catch (e) {
        entry.respError = e.message;
      }
    }
  }

  _onConsole(msg) {
    // Capture console errors only
    if (msg.type() === 'error' || msg.type() === 'warning') {
      this._console.push({ ts: Date.now(), type: msg.type(), text: msg.text(), page: this.currentPage?.id });
    }
  }
  _console = [];

  _onPageError(err) {
    this._console.push({ ts: Date.now(), type: 'pageerror', text: err.message, page: this.currentPage?.id });
  }

  _onFrameNavigated(frame) {
    if (frame === frame.page().mainFrame()) {
      if (this.verbose) console.log(`[nav] ${frame.url()}`);
    }
  }

  _findEntry(url, method) {
    return this.apiCalls.find(c => c.url === url && c.method === method && !c.status)
      || this.assets.find(c => c.url === url && c.method === method && !c.status);
  }

  summary() {
    const apiByPath = {};
    for (const c of this.apiCalls) {
      try {
        const u = new URL(c.url);
        const key = u.pathname;
        if (!apiByPath[key]) apiByPath[key] = { path: key, methods: new Set(), count: 0, sampleBody: null, sampleResponse: null, status: null };
        apiByPath[key].methods.add(c.method);
        apiByPath[key].count++;
        if (c.status) apiByPath[key].status = c.status;
        if (c.postData && !apiByPath[key].sampleBody) apiByPath[key].sampleBody = c.postData;
        if (c.respBody && !apiByPath[key].sampleResponse) apiByPath[key].sampleResponse = c.respBody.substring(0, 500);
      } catch (e) {}
    }
    return Object.values(apiByPath).map(e => ({ ...e, methods: [...e.methods] })).sort((a, b) => b.count - a.count);
  }
}
