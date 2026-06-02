// CLI argument parsing and orchestration
import { parseArgs } from 'node:util';
import { existsSync, mkdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { runMap } from './map.js';

function help() {
  console.log(`mimo-cli - end-to-end mapper for aistudio.xiaomimimo.com

Usage:
  mimo-cli map [options]

Options:
  --auth <file>          JSON file with { serviceToken, userId, xiaomichatbot_ph }
                         (or use MIMO_AUTH env var as JSON string)
  --out <dir>            Output directory (default: ./output)
  --base <url>           Base URL (default: https://aistudio.xiaomimimo.com)
  --routes <list>        Comma-separated routes to crawl (default: auto-discover)
  --max-pages <n>        Max pages to crawl (default: 30)
  --no-download          Skip downloading static assets
  --no-interact          Skip UI interaction (just load + map)
  --headless <bool>      Headless mode (default: true)
  --stealth <bool>       Use stealth plugin (default: true)
  --verbose              Verbose logging
  --help                 Show this help

Examples:
  MIMO_AUTH='{"serviceToken":"...","userId":"...","xiaomichatbot_ph":"..."}' mimo-cli map
  mimo-cli map --auth ./auth.json --out ./run1
`);
}

async function main() {
  const { values, positionals } = parseArgs({
    options: {
      auth: { type: 'string' },
      out: { type: 'string', default: './output' },
      base: { type: 'string', default: 'https://aistudio.xiaomimimo.com' },
      routes: { type: 'string' },
      'max-pages': { type: 'string', default: '30' },
      'no-download': { type: 'boolean', default: false },
      'no-interact': { type: 'boolean', default: false },
      headless: { type: 'string', default: 'true' },
      stealth: { type: 'string', default: 'true' },
      verbose: { type: 'boolean', default: false },
      help: { type: 'boolean', default: false },
    },
    allowPositionals: true,
  });

  if (values.help || (values._ && values._[0] === 'help')) {
    help();
    return;
  }

  const cmd = positionals[0] || 'map';
  if (cmd !== 'map') {
    console.error(`Unknown command: ${cmd}`);
    help();
    process.exit(1);
  }

  // Load auth
  let auth = null;
  if (values.auth) {
    if (!existsSync(values.auth)) {
      console.error(`Auth file not found: ${values.auth}`);
      process.exit(1);
    }
    auth = JSON.parse(readFileSync(values.auth, 'utf8'));
  } else if (process.env.MIMO_AUTH) {
    auth = JSON.parse(process.env.MIMO_AUTH);
  } else {
    console.error('No auth provided. Use --auth <file> or MIMO_AUTH env var.');
    help();
    process.exit(1);
  }

  for (const k of ['serviceToken', 'userId', 'xiaomichatbot_ph']) {
    if (!auth[k]) {
      console.error(`Auth missing required field: ${k}`);
      process.exit(1);
    }
  }

  const opts = {
    auth,
    outDir: values.out,
    base: values.base,
    routes: values.routes ? values.routes.split(',').map(r => r.trim()) : null,
    maxPages: parseInt(values['max-pages'], 10),
    download: !values['no-download'],
    interact: !values['no-interact'],
    headless: values.headless === 'true',
    stealth: values.stealth === 'true',
    verbose: values.verbose,
  };

  // Ensure out dir
  mkdirSync(opts.outDir, { recursive: true });
  mkdirSync(join(opts.outDir, 'assets'), { recursive: true });

  console.log('[mimo-cli] starting mapper');
  console.log(`[mimo-cli] base: ${opts.base}`);
  console.log(`[mimo-cli] out:  ${opts.outDir}`);
  console.log(`[mimo-cli] user: ${auth.userId}`);

  const t0 = Date.now();
  const result = await runMap(opts);
  const dt = ((Date.now() - t0) / 1000).toFixed(1);
  console.log(`[mimo-cli] done in ${dt}s`);
  console.log(`[mimo-cli] pages crawled: ${result.pages.length}`);
  console.log(`[mimo-cli] api calls:     ${result.apiCalls.length}`);
  console.log(`[mimo-cli] assets:        ${result.assets.length}`);
  console.log(`[mimo-cli] report:        ${join(opts.outDir, 'REPORT.md')}`);
}

main().catch(err => {
  console.error('[mimo-cli] fatal:', err);
  process.exit(1);
});
