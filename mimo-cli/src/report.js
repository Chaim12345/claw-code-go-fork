// Report generator
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';

export function writeJsonReport(outDir, data) {
  writeFileSync(join(outDir, 'api-map.json'), JSON.stringify(data.apiSummary, null, 2));
  writeFileSync(join(outDir, 'ui-map.json'), JSON.stringify(data.pages, null, 2));
  writeFileSync(join(outDir, 'all-api-calls.json'), JSON.stringify(data.apiCalls, null, 2));
  writeFileSync(join(outDir, 'all-assets.json'), JSON.stringify(data.assets, null, 2));
  writeFileSync(join(outDir, 'reverse-api.json'), JSON.stringify(data.reverseApi, null, 2));
  writeFileSync(join(outDir, 'asset-downloads.json'), JSON.stringify(data.assetDownloads, null, 2));
  writeFileSync(join(outDir, 'console-log.json'), JSON.stringify(data.consoleLog, null, 2));
}

export function writeMarkdownReport(outDir, data) {
  const lines = [];
  lines.push(`# Xiaomi MiMo Studio - Reverse Engineering Report`);
  lines.push('');
  lines.push(`Generated: ${new Date().toISOString()}`);
  lines.push(`Base URL: ${data.base}`);
  lines.push(`User: ${data.userId}`);
  lines.push('');
  lines.push(`## Summary`);
  lines.push(`- Pages crawled: **${data.pages.length}**`);
  lines.push(`- Unique API calls observed: **${data.apiSummary.length}**`);
  lines.push(`- Static assets captured: **${data.assets.length}**`);
  lines.push(`- Static assets downloaded: **${data.assetDownloads.filter(d => d.status === 200).length}**`);
  lines.push(`- API paths reverse-engineered from JS: **${data.reverseApi.length}**`);
  lines.push(`- Console errors: **${data.consoleLog.filter(c => c.type === 'pageerror').length}**`);
  lines.push('');

  // API summary
  lines.push(`## Internal API (observed)`);
  lines.push('');
  lines.push(`| Method | Path | Calls | Status |`);
  lines.push(`|---|---|---:|---|`);
  for (const e of data.apiSummary) {
    lines.push(`| ${e.methods.join(',')} | \`${e.path}\` | ${e.count} | ${e.status || '?'} |`);
  }
  lines.push('');

  // API sample request/response bodies
  lines.push(`## API Sample Request/Response Bodies`);
  lines.push('');
  for (const e of data.apiSummary.filter(x => x.sampleBody || x.sampleResponse)) {
    lines.push(`### \`${e.path}\``);
    if (e.sampleBody) {
      lines.push('**Request body sample:**');
      lines.push('```json');
      lines.push(truncate(e.sampleBody, 1500));
      lines.push('```');
    }
    if (e.sampleResponse) {
      lines.push('**Response body sample:**');
      lines.push('```json');
      lines.push(truncate(e.sampleResponse, 1500));
      lines.push('```');
    }
    if (e.sseEvents?.length) {
      lines.push(`**SSE events:** ${e.sseEvents.join(', ')}`);
    }
    lines.push('');
  }

  // Reverse-engineered paths
  lines.push(`## API Paths Discovered in JS Bundles (static analysis)`);
  lines.push('');
  lines.push(`| Method | Path | Occurrences |`);
  lines.push(`|---|---|---:|`);
  for (const e of data.reverseApi.slice(0, 100)) {
    lines.push(`| ${e.method || '?'} | \`${e.path}\` | ${e.occurrences} |`);
  }
  lines.push('');

  // Pages
  lines.push(`## Pages Discovered`);
  lines.push('');
  for (const p of data.pages) {
    lines.push(`### ${p.route || '/'} (HTTP ${p.status || '?'})`);
    if (p.snapshot) {
      lines.push(`- Title: ${p.snapshot.title}`);
      if (p.snapshot.h1.length) lines.push(`- Headings: ${p.snapshot.h1.join(' | ')}`);
      if (p.snapshot.componentCounts?.length) {
        const cc = p.snapshot.componentCounts.slice(0, 5).map(([k, v]) => `${k}=${v}`).join(', ');
        lines.push(`- Components: ${cc}`);
      }
      lines.push(`- Interactive elements: ${p.snapshot.interactive.length}`);
      lines.push(`- Forms/inputs: ${p.snapshot.forms.length}`);
    }
    lines.push(`- API calls on this page: ${p.api.length}`);
    lines.push(`- Assets loaded: ${p.assets.length}`);
    if (p.snapshot?.bodyText) {
      lines.push('');
      lines.push('> ' + p.snapshot.bodyText.substring(0, 500).replace(/\n+/g, ' '));
    }
    lines.push('');
  }

  // Asset breakdown
  lines.push(`## Static Asset Breakdown`);
  lines.push('');
  const byType = {};
  for (const a of data.assets) {
    const ext = (a.url.match(/\.([a-zA-Z0-9]{1,6})(?:\?|$)/) || [,'unknown'])[1].toLowerCase();
    byType[ext] = (byType[ext] || 0) + 1;
  }
  lines.push(`| Type | Count |`);
  lines.push(`|---|---:|`);
  Object.entries(byType).sort((a, b) => b[1] - a[1]).forEach(([k, v]) => lines.push(`| .${k} | ${v} |`));
  lines.push('');

  writeFileSync(join(outDir, 'REPORT.md'), lines.join('\n'));
}

function truncate(s, n) {
  if (!s) return '';
  return s.length > n ? s.substring(0, n) + '\n...[truncated]' : s;
}
