// Static asset downloader
import { createWriteStream, existsSync, mkdirSync, statSync } from 'node:fs';
import { writeFile } from 'node:fs/promises';
import { dirname, join, basename } from 'node:path';
import { pipeline } from 'node:stream/promises';
import { Readable } from 'node:stream';

export class AssetDownloader {
  constructor({ outDir, verbose = false }) {
    this.outDir = outDir;
    this.verbose = verbose;
    this.downloaded = new Set();
  }

  // Compute a safe local path mirroring the URL path
  localPathFor(urlStr) {
    const u = new URL(urlStr);
    let p = u.pathname;
    if (p.endsWith('/')) p += 'index.html';
    if (!/\.[a-zA-Z0-9]{1,6}$/.test(basename(p))) p += '.bin';
    // Strip leading slash
    return p.replace(/^\/+/, '');
  }

  async downloadOne(urlStr) {
    if (this.downloaded.has(urlStr)) return null;
    this.downloaded.add(urlStr);
    const rel = this.localPathFor(urlStr);
    const dest = join(this.outDir, 'assets', rel);
    mkdirSync(dirname(dest), { recursive: true });
    try {
      const res = await fetch(urlStr, { redirect: 'follow' });
      if (!res.ok) {
        return { url: urlStr, local: rel, status: res.status, error: 'http ' + res.status };
      }
      const buf = Buffer.from(await res.arrayBuffer());
      await writeFile(dest, buf);
      return { url: urlStr, local: rel, status: 200, size: buf.length };
    } catch (e) {
      return { url: urlStr, local: rel, error: e.message };
    }
  }

  async downloadAll(assets) {
    const results = [];
    // Dedupe by URL
    const unique = [...new Set(assets.map(a => a.url))];
    if (this.verbose) console.log(`[assets] downloading ${unique.length} unique files`);
    let i = 0;
    for (const url of unique) {
      if (i++ % 10 === 0 && this.verbose) console.log(`[assets] ${i}/${unique.length}`);
      results.push(await this.downloadOne(url));
    }
    return results.filter(Boolean);
  }
}
