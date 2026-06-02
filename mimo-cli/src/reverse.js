// API reverse engineering - extract endpoint definitions from JS bundles
import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';

const API_PATH_PATTERNS = [
  // /open-apis/...
  /["'`]\/open-apis\/[a-zA-Z0-9_\-/{}:.]+["'`]/g,
  // /api/...
  /["'`]\/api\/[a-zA-Z0-9_\-/{}:.]+["'`]/g,
];

const HTTP_METHOD_KEYWORDS = ['get', 'post', 'put', 'delete', 'patch'];

export class ApiReverser {
  constructor({ assetsDir, verbose = false }) {
    this.assetsDir = assetsDir;
    this.verbose = verbose;
    this.endpoints = new Map();
  }

  walk(dir) {
    const out = [];
    const stack = [dir];
    while (stack.length) {
      const cur = stack.pop();
      let stat;
      try { stat = statSync(cur); } catch (e) { continue; }
      if (stat.isDirectory()) {
        for (const ent of readdirSync(cur)) stack.push(join(cur, ent));
      } else if (cur.endsWith('.js')) {
        out.push(cur);
      }
    }
    return out;
  }

  extractFromFile(file) {
    let content;
    try { content = readFileSync(file, 'utf8'); } catch (e) { return []; }
    if (content.length < 100) return [];

    const found = [];
    for (const pat of API_PATH_PATTERNS) {
      const matches = content.match(pat) || [];
      for (const m of matches) {
        const path = m.slice(1, -1);
        if (path.length < 4) continue;
        found.push(path);
      }
    }
    return found;
  }

  // Find context around each path: method (POST/GET), body shape, query params
  extractContext(file) {
    let content;
    try { content = readFileSync(file, 'utf8'); } catch (e) { return []; }
    const out = [];
    for (const pat of API_PATH_PATTERNS) {
      const re = new RegExp(pat.source, 'g');
      let m;
      while ((m = re.exec(content)) !== null) {
        const path = m[0].slice(1, -1);
        if (path.length < 4) continue;
        // Look back 300 chars for method hint
        const ctx = content.substring(Math.max(0, m.index - 400), m.index);
        const methodMatch = ctx.match(/["'`](get|post|put|delete|patch)["'`]/i);
        const method = methodMatch ? methodMatch[1].toUpperCase() : null;
        out.push({ path, method, ctxSnippet: ctx.substring(Math.max(0, ctx.length - 200)) });
      }
    }
    return out;
  }

  reverse() {
    const files = this.walk(this.assetsDir);
    if (this.verbose) console.log(`[reverse] scanning ${files.length} JS files`);
    for (const f of files) {
      const ctxs = this.extractContext(f);
      for (const c of ctxs) {
        const key = c.method ? `${c.method} ${c.path}` : c.path;
        if (!this.endpoints.has(key)) {
          this.endpoints.set(key, { method: c.method, path: c.path, sources: [], occurrences: 0 });
        }
        const e = this.endpoints.get(key);
        e.occurrences++;
        e.sources.push(f.replace(this.assetsDir + '/', ''));
      }
    }
    return [...this.endpoints.values()].sort((a, b) => b.occurrences - a.occurrences);
  }
}
