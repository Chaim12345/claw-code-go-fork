// Main orchestrator
import { launchBrowser, newPage } from './browser.js';
import { ApiRecorder } from './recorder.js';
import { UiMapper } from './ui.js';
import { AssetDownloader } from './assets.js';
import { ApiReverser } from './reverse.js';
import { writeJsonReport, writeMarkdownReport } from './report.js';

export async function runMap(opts) {
  const browser = await launchBrowser({ headless: opts.headless, stealth: opts.stealth });
  const page = await newPage(browser, { stealth: opts.stealth, cookies: opts.auth, baseUrl: opts.base, verbose: opts.verbose });
  const recorder = new ApiRecorder({ verbose: opts.verbose });
  recorder.attach(page);

  // 1. Crawl UI
  const mapper = new UiMapper({ base: opts.base, routes: opts.routes, maxPages: opts.maxPages, interact: opts.interact, verbose: opts.verbose });
  const pages = await mapper.crawl(page, recorder);

  // 2. Reverse engineer API from JS bundles
  const reverser = new ApiReverser({ assetsDir: `${opts.outDir}/assets`, verbose: opts.verbose });
  const reverseApi = reverser.reverse();

  // 3. Download all static assets
  let assetDownloads = [];
  if (opts.download) {
    const downloader = new AssetDownloader({ outDir: opts.outDir, verbose: opts.verbose });
    assetDownloads = await downloader.downloadAll(recorder.assets);
  }

  await browser.close();

  // 4. Build summary
  const apiSummary = recorder.summary();
  const data = {
    base: opts.base,
    userId: opts.auth.userId,
    pages,
    apiCalls: recorder.apiCalls,
    assets: recorder.assets,
    apiSummary,
    reverseApi,
    assetDownloads,
    consoleLog: recorder._console,
  };

  writeJsonReport(opts.outDir, data);
  writeMarkdownReport(opts.outDir, data);
  return data;
}
