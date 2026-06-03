/**
 * claw-code-go chat UI — client-side JavaScript
 * Mobile-first, vanilla JS. Handles WebSocket connection,
 * message rendering, virtual keyboard, tool cards, permissions.
 */

(function () {
  'use strict';

  var composer       = document.getElementById('composer');
  var composerInput  = document.getElementById('composer-input');
  var sendButton     = document.getElementById('send-button');
  var charCountEl    = document.getElementById('char-count');
  var messagesEl     = document.getElementById('messages');
  var statusPill     = document.getElementById('connection-status');
  var sidebarToggle  = document.getElementById('sidebar-toggle');
  var sidebar        = document.getElementById('sidebar');

  var ws = null;
  var reconnectTimer = null;
  var reconnectDelay = 1000;
  var MAX_RECONNECT_DELAY = 30000;
  var currentBubble = null;
  var currentBubbleContent = '';
  var toolStartTime = null;
  var pendingToolCard = null;

  // Reconnection / dedup state.
  var pendingMessages = {};    // message_id → {text, timestamp}
  var ackedMessages = {};      // message_id → true (pruned after 60s)

  // ── Theme switcher (system / light / dark) ───────────────
  var themeToggle   = document.getElementById('theme-toggle');
  var THEME_KEY     = 'claw_theme';
  var THEME_CYCLE   = ['system', 'light', 'dark'];

  function getCurrentTheme() {
    return document.documentElement.getAttribute('data-theme') || 'system';
  }

  function applyTheme(theme) {
    if (theme === 'system') {
      document.documentElement.removeAttribute('data-theme');
    } else {
      document.documentElement.setAttribute('data-theme', theme);
    }
    localStorage.setItem(THEME_KEY, theme);
    updateThemeToggleLabel(theme);
  }

  function updateThemeToggleLabel(theme) {
    if (!themeToggle) return;
    var labels = { system: '◐', light: '☀', dark: '☾' };
    themeToggle.textContent = labels[theme] || '◐';
  }

  function cycleTheme() {
    var current = getCurrentTheme();
    var idx = THEME_CYCLE.indexOf(current);
    var next = THEME_CYCLE[(idx + 1) % THEME_CYCLE.length];
    applyTheme(next);
  }

  // Restore persisted theme on load.
  (function () {
    var saved = localStorage.getItem(THEME_KEY);
    if (saved && (saved === 'light' || saved === 'dark')) {
      applyTheme(saved);
    } else if (saved === 'system') {
      // Clear any explicit attribute so prefers-color-scheme takes over.
      document.documentElement.removeAttribute('data-theme');
      updateThemeToggleLabel('system');
    }
  })();

  if (themeToggle) {
    themeToggle.addEventListener('click', cycleTheme);
  }

  // ── Font size control (S / M / L) ────────────────────────
  var fontSizeToggle = document.getElementById('font-size-toggle');
  var FONT_SIZE_KEY  = 'claw_font_size';
  var FONT_SIZE_CYCLE = ['s', 'm', 'l'];

  function getCurrentFontSize() {
    return document.documentElement.getAttribute('data-font-size') || 'm';
  }

  function applyFontSize(size) {
    if (size === 'm') {
      document.documentElement.removeAttribute('data-font-size');
    } else {
      document.documentElement.setAttribute('data-font-size', size);
    }
    localStorage.setItem(FONT_SIZE_KEY, size);
    updateFontSizeToggleLabel(size);
  }

  function updateFontSizeToggleLabel(size) {
    if (!fontSizeToggle) return;
    var labels = { s: 'S', m: 'M', l: 'L' };
    fontSizeToggle.textContent = labels[size] || 'A';
    fontSizeToggle.style.fontWeight = size === 'l' ? '700' : '500';
    fontSizeToggle.style.fontSize = size === 's' ? '0.7rem' : size === 'l' ? '0.95rem' : '0.85rem';
  }

  function cycleFontSize() {
    var current = getCurrentFontSize();
    var idx = FONT_SIZE_CYCLE.indexOf(current);
    var next = FONT_SIZE_CYCLE[(idx + 1) % FONT_SIZE_CYCLE.length];
    applyFontSize(next);
  }

  // Restore persisted font size on load.
  (function () {
    var saved = localStorage.getItem(FONT_SIZE_KEY);
    if (saved === 's' || saved === 'l') {
      applyFontSize(saved);
    } else {
      // Default: medium (no attribute)
      updateFontSizeToggleLabel('m');
    }
  })();

  if (fontSizeToggle) {
    fontSizeToggle.addEventListener('click', cycleFontSize);
  }

  // ── UUID generation (for message dedup) ──────────────────
  function generateUUID() {
    // crypto.randomUUID() is available in all modern browsers.
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
      return crypto.randomUUID();
    }
    // Fallback for older browsers.
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function (c) {
      var r = Math.random() * 16 | 0;
      var v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  }

  // ── Visual Viewport ──────────────────────────────────────
  function handleViewportResize() {
    if (!window.visualViewport) return;
    var vp = window.visualViewport;
    var kh = window.innerHeight - vp.height;
    if (kh > 100) { composer.style.paddingBottom = (kh + 8) + 'px'; scrollToBottom(); }
    else { composer.style.paddingBottom = ''; }
    if (vp.offsetTop > 0) messagesEl.scrollTop += vp.offsetTop;
  }
  function setupKeyboardHandling() {
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', handleViewportResize);
      window.visualViewport.addEventListener('scroll', handleViewportResize);
    }
  }

  // ── Composer ─────────────────────────────────────────────
  var CHAR_COUNT_WARN_THRESHOLD = 0.9;   // show when > 90% of maxlength
  var AUTO_GROW_MAX_ROWS = 8;

  function autoGrowTextarea() {
    composerInput.style.height = 'auto';
    var lh = parseFloat(getComputedStyle(composerInput).lineHeight) || 24;
    var maxHeight = lh * AUTO_GROW_MAX_ROWS +
                    parseFloat(getComputedStyle(composerInput).paddingTop || 0) +
                    parseFloat(getComputedStyle(composerInput).paddingBottom || 0);
    composerInput.style.height = Math.min(composerInput.scrollHeight, maxHeight) + 'px';
    // Toggle overflow for very long text beyond max rows
    composerInput.style.overflowY = composerInput.scrollHeight > maxHeight ? 'auto' : 'hidden';
  }

  function updateCharCount() {
    if (!charCountEl) return;
    var len = composerInput.value.length;
    var max = parseInt(composerInput.getAttribute('maxlength'), 10) || 100000;
    var threshold = max * CHAR_COUNT_WARN_THRESHOLD;

    if (len > threshold) {
      charCountEl.textContent = len.toLocaleString() + ' / ' + max.toLocaleString();
      charCountEl.className = len > max * 0.98 ? 'char-count-warn' : 'char-count-visible';
    } else {
      charCountEl.className = 'char-count-hidden';
    }
  }

  function sendMessage() {
    var text = composerInput.value.trim();
    if (!text || !ws || ws.readyState !== WebSocket.OPEN) return;
    var msgId = generateUUID();
    var payload = { type: 'user_input', text: text, message_id: msgId };
    pendingMessages[msgId] = { text: text, timestamp: Date.now() };
    ws.send(JSON.stringify(payload));
    appendMessage('user', escapeHtml(text));
    // Show thinking indicator after a delay (canceled when response arrives).
    showThinkingIndicator();
    composerInput.value = '';
    composerInput.style.height = '';
    composerInput.style.overflowY = 'hidden';
    sendButton.disabled = true;
    updateCharCount();
    composerInput.focus();
    scrollToBottom();
  }

  composerInput.addEventListener('input', function () {
    autoGrowTextarea();
    updateCharCount();
    sendButton.disabled = this.value.trim().length === 0;
    if (window.visualViewport) handleViewportResize();
  });

  // Enter to submit, Shift+Enter for newline
  composerInput.addEventListener('keydown', function (e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
    // Shift+Enter: let default behavior insert newline (no-op here)
  });

  // Still handle form submit for the send button click
  composer.addEventListener('submit', function (e) {
    e.preventDefault();
    sendMessage();
  });

  // ── Message rendering ────────────────────────────────────
  // ── Empty state management ────────────────────────────
  var emptyStateEl = document.getElementById('empty-state');

  function showEmptyState() {
    if (emptyStateEl) emptyStateEl.style.display = '';
  }

  function hideEmptyState() {
    if (emptyStateEl) emptyStateEl.style.display = 'none';
  }

  // Check whether any real messages exist (excluding empty-state and thinking).
  function hasMessages() {
    var children = messagesEl.children;
    for (var i = 0; i < children.length; i++) {
      var c = children[i];
      if (c.id === 'empty-state') continue;
      if (c.classList.contains('thinking-indicator')) continue;
      return true;
    }
    return false;
  }

  function updateEmptyState() {
    if (hasMessages()) {
      hideEmptyState();
    } else {
      showEmptyState();
    }
  }

  // Initial state: show empty state (no messages yet).
  updateEmptyState();

  function appendMessage(role, html) {
    var wrapper = document.createElement('div');
    wrapper.className = 'message message-' + role;
    wrapper.innerHTML = html;
    messagesEl.appendChild(wrapper);
    hideEmptyState();
    scrollToBottom();
  }

  function startAssistantBubble() {
    currentBubble = document.createElement('div');
    currentBubble.className = 'message message-assistant';
    currentBubbleContent = '';
    messagesEl.appendChild(currentBubble);
  }

  function appendDelta(text) {
    if (!currentBubble) startAssistantBubble();
    currentBubbleContent += text;
    currentBubble.innerHTML = renderMarkdown(currentBubbleContent);
    scrollToBottom();
  }

  function finalizeAssistantBubble() {
    if (currentBubble) {
      currentBubble.innerHTML = renderMarkdown(currentBubbleContent);
      enhanceCodeBlocks(currentBubble);
      currentBubble = null;
      currentBubbleContent = '';
    }
    // Remove thinking indicator when text arrives.
    removeThinkingIndicator();
  }

  // ── Thinking indicator (shows "Thinking…" with elapsed time) ────
  var thinkingEl = null;
  var thinkingStartTime = 0;
  var thinkingTimerInterval = null;
  var THINKING_SHOW_DELAY = 1500; // show after 1.5s of waiting

  function showThinkingIndicator() {
    // Don't show if we already have one.
    if (thinkingEl) return;

    // Wait THINKING_SHOW_DELAY ms before showing — if we get a response
    // before then, cancel.
    thinkingStartTime = Date.now();
    setTimeout(function () {
      // Check that we still haven't gotten a response.
      if (thinkingEl) return;
      if (!thinkingStartTime) return; // cancelled
      if (currentBubble && currentBubbleContent) return; // already getting text

      thinkingEl = document.createElement('div');
      thinkingEl.className = 'thinking-indicator';
      thinkingEl.setAttribute('aria-label', 'Thinking…');

      var dots = document.createElement('span');
      dots.className = 'thinking-dots';
      dots.innerHTML = '\u003cspan\u003e\u003c/span\u003e\u003cspan\u003e\u003c/span\u003e\u003cspan\u003e\u003c/span\u003e';
      thinkingEl.appendChild(dots);

      var label = document.createElement('span');
      label.textContent = 'Thinking';
      thinkingEl.appendChild(label);

      var timer = document.createElement('span');
      timer.className = 'thinking-timer';
      thinkingEl.appendChild(timer);

      messagesEl.appendChild(thinkingEl);
      hideEmptyState();
      scrollToBottom();

      // Update elapsed time every second.
      clearInterval(thinkingTimerInterval);
      thinkingTimerInterval = setInterval(function () {
        if (!thinkingEl) { clearInterval(thinkingTimerInterval); return; }
        var elapsed = Math.floor((Date.now() - thinkingStartTime) / 1000);
        timer.textContent = elapsed + 's';
      }, 1000);
    }, THINKING_SHOW_DELAY);
  }

  function removeThinkingIndicator() {
    thinkingStartTime = 0;
    clearInterval(thinkingTimerInterval);
    thinkingTimerInterval = null;
    if (thinkingEl) {
      thinkingEl.remove();
      thinkingEl = null;
    }
  }

  // ── Tool-call cards ──────────────────────────────────────
  function createToolCard(toolName, inputSummary) {
    var card = document.createElement('div');
    card.className = 'message message-tool';

    var summary = document.createElement('div');
    summary.className = 'tool-summary';

    var icon = document.createElement('span');
    icon.className = 'tool-icon';
    icon.textContent = '\u2699';
    summary.appendChild(icon);

    var nameEl = document.createElement('strong');
    nameEl.textContent = escapeHtml(toolName || 'tool');
    summary.appendChild(nameEl);

    if (inputSummary) {
      var desc = document.createElement('span');
      desc.className = 'tool-desc';
      desc.textContent = ' — ' + escapeHtml(inputSummary);
      summary.appendChild(desc);
    }

    var timer = document.createElement('span');
    timer.className = 'tool-timer';
    timer.textContent = '0.0s';
    summary.appendChild(timer);
    card.appendChild(summary);

    var details = document.createElement('details');
    details.className = 'tool-details';
    var dtSummary = document.createElement('summary');
    dtSummary.textContent = 'result';
    details.appendChild(dtSummary);

    var pre = document.createElement('pre');
    var code = document.createElement('code');
    code.className = 'tool-output';
    code.textContent = '';
    pre.appendChild(code);
    details.appendChild(pre);

    var copyBtn = document.createElement('button');
    copyBtn.className = 'tool-copy-btn';
    copyBtn.textContent = 'Copy';
    copyBtn.addEventListener('click', function () {
      navigator.clipboard.writeText(code.textContent).then(function () {
        copyBtn.textContent = 'Copied!';
        setTimeout(function () { copyBtn.textContent = 'Copy'; }, 2000);
      }).catch(function () { copyBtn.textContent = 'Error'; });
    });
    details.appendChild(copyBtn);
    card.appendChild(details);
    messagesEl.appendChild(card);

    return { card: card, code: code, timer: timer };
  }

  function startToolCard(toolName, inputSummary) {
    finalizeAssistantBubble();
    toolStartTime = Date.now();
    var result = createToolCard(toolName, inputSummary);
    pendingToolCard = result;
    var updateTimer = function () {
      if (pendingToolCard !== result) return;
      var elapsed = ((Date.now() - toolStartTime) / 1000).toFixed(1);
      result.timer.textContent = elapsed + 's';
      if (pendingToolCard === result) {
        requestAnimationFrame(function () { setTimeout(updateTimer, 100); });
      }
    };
    updateTimer();
  }

  function finishToolCard(output) {
    if (pendingToolCard) {
      pendingToolCard.code.textContent = String(output || '').substring(0, 2000);
      pendingToolCard = null;
      toolStartTime = null;
    }
  }

  // ── Permission prompts ───────────────────────────────────
  function showPermissionDialog(msg) {
    finalizeAssistantBubble();

    var overlay = document.createElement('div');
    overlay.className = 'perm-overlay';

    var card = document.createElement('div');
    card.className = 'perm-card';

    var title = document.createElement('h3');
    title.textContent = 'Permission Request';
    card.appendChild(title);

    var toolName = document.createElement('div');
    toolName.className = 'perm-tool-name';
    toolName.textContent = escapeHtml(msg.tool_name || 'unknown tool');
    card.appendChild(toolName);

    if (msg.tool_input) {
      var input = document.createElement('pre');
      input.className = 'perm-input';
      input.textContent = escapeHtml(String(msg.tool_input).substring(0, 500));
      card.appendChild(input);
    }

    var actions = document.createElement('div');
    actions.className = 'perm-actions';

    function reply(decision) {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'permission_reply',
          tool_use_id: msg.tool_use_id || '',
          decision: decision
        }));
      }
      overlay.remove();
    }

    var allowOnce = document.createElement('button');
    allowOnce.className = 'perm-btn perm-allow-once';
    allowOnce.textContent = 'Allow once';
    allowOnce.addEventListener('click', function () { reply('allow'); });

    var allowAlways = document.createElement('button');
    allowAlways.className = 'perm-btn perm-allow-always';
    allowAlways.textContent = 'Always allow';
    allowAlways.addEventListener('click', function () { reply('allow_always'); });

    var denyBtn = document.createElement('button');
    denyBtn.className = 'perm-btn perm-deny';
    denyBtn.textContent = 'Deny';
    denyBtn.addEventListener('click', function () { reply('deny'); });

    actions.appendChild(allowOnce);
    actions.appendChild(allowAlways);
    actions.appendChild(denyBtn);
    card.appendChild(actions);

    overlay.appendChild(card);
    document.body.appendChild(overlay);
  }

  // ── Markdown renderer ────────────────────────────────────
  function renderMarkdown(text) {
    var escaped = escapeHtml(text);
    escaped = escaped.replace(/```(\w+)?\n([\s\S]*?)```/g, function (match, lang, code) {
      var langClass = lang ? ' class="language-' + lang + '"' : '';
      return '<pre><code' + langClass + '>' + code + '</code></pre>';
    });
    escaped = escaped.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>');
    escaped = escaped.replace(/\n/g, '<br>');
    return escaped;
  }

  function escapeHtml(str) {
    return str
      .replace(/&/g, '&')
      .replace(/</g, '<')
      .replace(/>/g, '>')
      .replace(/"/g, '"');
  }

  // ── Code block actions (copy, download, run in terminal) ─
  function enhanceCodeBlocks(bubble) {
    var pres = bubble.querySelectorAll('pre');
    for (var i = 0; i < pres.length; i++) {
      var pre = pres[i];
      // Skip if already enhanced.
      if (pre.parentNode.classList.contains('code-block-wrapper')) continue;

      var code = pre.querySelector('code');
      if (!code) continue;
      var codeText = code.textContent || '';

      // Wrap pre in a container for the action bar.
      var wrapper = document.createElement('div');
      wrapper.className = 'code-block-wrapper';
      pre.parentNode.insertBefore(wrapper, pre);
      wrapper.appendChild(pre);

      // Determine language from class e.g. "language-go".
      var lang = '';
      if (code.className) {
        var match = code.className.match(/language-(\w+)/);
        if (match) lang = match[1];
      }

      // Detect filename from a leading comment line.
      var filename = detectFilename(codeText, lang);
      if (filename) {
        var fnEl = document.createElement('span');
        fnEl.className = 'code-filename';
        fnEl.textContent = filename;
        wrapper.appendChild(fnEl);
      }

      // Action bar with buttons.
      var bar = document.createElement('div');
      bar.className = 'code-actions';

      // Copy button.
      var copyBtn = document.createElement('button');
      copyBtn.className = 'code-action-btn';
      copyBtn.textContent = 'Copy';
      copyBtn.title = 'Copy to clipboard';
      copyBtn.addEventListener('click', function () {
        copyCodeBlock(codeText, copyBtn);
      });
      bar.appendChild(copyBtn);

      // Download button.
      var dlBtn = document.createElement('button');
      dlBtn.className = 'code-action-btn';
      dlBtn.textContent = 'Download';
      dlBtn.title = 'Download as file';
      dlBtn.addEventListener('click', function () {
        downloadCodeBlock(codeText, filename, lang);
      });
      bar.appendChild(dlBtn);

      // "Run in terminal" button — only if a TUI session is available.
      // The /terminal page uses the same host for wterm, so we check if
      // the user could plausibly have a terminal session open.
      var runBtn = document.createElement('button');
      runBtn.className = 'code-action-btn code-action-run';
      runBtn.textContent = '▶ Run';
      runBtn.title = 'Run in terminal (open /terminal first)';
      runBtn.addEventListener('click', function () {
        runInTerminal(codeText, lang);
      });
      bar.appendChild(runBtn);

      wrapper.appendChild(bar);
    }
  }

  function copyCodeBlock(text, btn) {
    navigator.clipboard.writeText(text).then(function () {
      btn.textContent = 'Copied!';
      setTimeout(function () { btn.textContent = 'Copy'; }, 2000);
    }).catch(function () {
      btn.textContent = 'Error';
      setTimeout(function () { btn.textContent = 'Copy'; }, 2000);
    });
  }

  function downloadCodeBlock(text, filename, lang) {
    var resolvedName = filename || 'code-snippet';
    // Add extension based on language if no filename provided.
    if (!filename && lang) {
      var ext = extensionForLang(lang);
      if (ext) resolvedName = resolvedName + ext;
    }
    var blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
    var url = URL.createObjectURL(blob);
    var a = document.createElement('a');
    a.href = url;
    a.download = resolvedName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  function runInTerminal(text, lang) {
    // Open /terminal in a new tab, then the user can paste the code.
    // We also provide a copy-to-clipboard convenience so the user
    // can paste immediately.
    var cmd = '';
    // For bash/shell snippets, the code is directly runnable.
    if (lang === 'bash' || lang === 'sh' || lang === 'shell' || !lang) {
      cmd = text;
    } else {
      // For other languages, wrap in a comment showing what to do.
      cmd = '# Copy the code below and run it in your terminal:\n' + text;
    }
    navigator.clipboard.writeText(cmd).then(function () {
      var win = window.open('/terminal', '_blank');
      if (!win) {
        // Popup blocked — just inform the user.
        appendMessage('assistant', 'Code copied to clipboard. <a href="/terminal" target="_blank" rel="noopener">Open terminal</a> to paste and run.');
      } else {
        // Brief toast-like feedback.
        appendMessage('assistant', 'Code copied to clipboard — paste it in the terminal tab.');
      }
    }).catch(function () {
      window.open('/terminal', '_blank');
    });
  }

  function extensionForLang(lang) {
    var map = {
      'go':         '.go',
      'golang':     '.go',
      'typescript': '.ts',
      'ts':         '.ts',
      'javascript': '.js',
      'js':         '.js',
      'python':     '.py',
      'py':         '.py',
      'bash':       '.sh',
      'sh':         '.sh',
      'shell':      '.sh',
      'json':       '.json',
      'html':       '.html',
      'css':        '.css',
      'yaml':       '.yaml',
      'yml':        '.yml',
      'toml':       '.toml',
      'rust':       '.rs',
      'rs':         '.rs',
      'c':          '.c',
      'cpp':        '.cpp',
      'java':       '.java',
      'sql':        '.sql',
      'dockerfile': '',
      'makefile':   ''
    };
    return map[lang] || '.' + lang;
  }

  // Detect a filename hint from a leading comment in the code.
  // e.g. "// foo/bar.go" or "# path/to/file.py" or "// File: main.go"
  function detectFilename(codeText, lang) {
    var lines = codeText.split('\n');
    if (!lines.length) return '';

    // Check first 3 lines for filename patterns.
    for (var i = 0; i < Math.min(lines.length, 3); i++) {
      var line = lines[i].trim();

      // Match: // path/to/file.go or # path/to/file.py or -- path/to/file.lua
      // The path must contain a dot (extension) and look like a file path.
      var commentPrefixes = ['//', '#', '--', ';;'];
      for (var j = 0; j < commentPrefixes.length; j++) {
        var prefix = commentPrefixes[j];
        if (line.indexOf(prefix) === 0) {
          var after = line.substring(prefix.length).trim();
          // Remove "File:" or "file:" prefix.
          after = after.replace(/^(?:File|file):\s*/i, '');
          // Check if it looks like a filename (has extension, no spaces).
          if (/^[\w./-]+\.[a-zA-Z]\w*$/.test(after)) {
            // Extract just the basename.
            var parts = after.split('/');
            return parts[parts.length - 1];
          }
          // Also match patterns like "filename: path/to/file" or "@file path/to/file"
        }
      }

      // Match: @file path/to/file.go
      var fileMatch = line.match(/^@file\s+([\w./-]+\.[a-zA-Z]\w*)/);
      if (fileMatch) {
        var fparts = fileMatch[1].split('/');
        return fparts[fparts.length - 1];
      }
    }

    // Heuristic: if code looks like a complete file, try to infer name.
    // e.g. "package main" → main.go
    if (!lang || lang === 'go') {
      var pkgMatch = codeText.match(/^package\s+(\w+)/m);
      if (pkgMatch) return pkgMatch[1] + '.go';
    }

    return '';
  }

  function scrollToBottom() {
    requestAnimationFrame(function () {
      messagesEl.scrollTop = messagesEl.scrollHeight;
    });
  }

  // ── Connection ───────────────────────────────────────────
  // ── Rate-limit banner ─────────────────────────────────
  var rateLimitBanner = document.getElementById('rate-limit-banner');
  var rateLimitCountdownEl = document.getElementById('rate-limit-countdown');
  var rateLimitCountdownTimer = null;

  function showRateLimitBanner(message) {
    if (!rateLimitBanner) return;
    rateLimitBanner.classList.add('rate-limit-visible');
    // Countdown from 30 seconds.
    var remaining = 30;
    if (rateLimitCountdownEl) rateLimitCountdownEl.textContent = remaining + 's';
    clearInterval(rateLimitCountdownTimer);
    rateLimitCountdownTimer = setInterval(function () {
      remaining--;
      if (remaining <= 0) {
        clearInterval(rateLimitCountdownTimer);
        rateLimitCountdownTimer = null;
        hideRateLimitBanner();
        return;
      }
      if (rateLimitCountdownEl) rateLimitCountdownEl.textContent = remaining + 's';
    }, 1000);
  }

  function hideRateLimitBanner() {
    if (!rateLimitBanner) return;
    rateLimitBanner.classList.remove('rate-limit-visible');
    clearInterval(rateLimitCountdownTimer);
    rateLimitCountdownTimer = null;
  }

  function connect() {
    var protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    var wsURL = protocol + '//' + window.location.host + '/api/chat/ws';
    var params = new URLSearchParams(window.location.search);
    var token = params.get('token');
    if (token) wsURL += '?token=' + encodeURIComponent(token);

    ws = new WebSocket(wsURL);

    ws.onopen = function () {
      setStatus('connected');
      reconnectDelay = 1000;
      hideRateLimitBanner();
      if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null; }
      // Resend any user_input messages that haven't been ack'd yet.
      var now = Date.now();
      var ids = Object.keys(pendingMessages);
      for (var i = 0; i < ids.length; i++) {
        var mid = ids[i];
        var pm = pendingMessages[mid];
        // Only resend messages sent in the last 2 minutes.
        if (now - pm.timestamp < 120000) {
          ws.send(JSON.stringify({ type: 'user_input', text: pm.text, message_id: mid }));
        } else {
          delete pendingMessages[mid];
        }
      }
    };

    ws.onmessage = function (event) {
      try {
        var msg = JSON.parse(event.data);
        switch (msg.type) {
          case 'chat_session_init':
            currentSessionID = msg.session_id || null;
            break;
          case 'ack':
            // Server acknowledged a user_input; clear it from pending.
            if (msg.message_id) {
              delete pendingMessages[msg.message_id];
              ackedMessages[msg.message_id] = true;
              // Prune old acked entries after 60 seconds.
              setTimeout(function () { delete ackedMessages[msg.message_id]; }, 60000);
            }
            break;
          case 'text_delta':  appendDelta(msg.text || ''); break;
          case 'text_final':  finalizeAssistantBubble(); break;
          case 'tool_start':  startToolCard(msg.tool_name, msg.input_summary || ''); break;
          case 'tool_done':   finishToolCard(msg.output || ''); break;
          case 'permission_ask': showPermissionDialog(msg); break;
          case 'error':
            finalizeAssistantBubble();
            if (msg.code === 'rate_limit') {
              showRateLimitBanner(msg.message || 'Rate limited');
            } else {
              appendMessage('error', 'Error: ' + escapeHtml(msg.message || msg.error || 'unknown'));
            }
            break;
          case 'done': finalizeAssistantBubble(); break;
          default: break;
        }
      } catch (err) { /* ignore non-JSON */ }
    };

    ws.onclose = function () { setStatus('disconnected'); scheduleReconnect(); };
    ws.onerror = function () {};
  }

  function scheduleReconnect() {
    if (reconnectTimer) return;
    setStatus('reconnecting');
    reconnectTimer = setTimeout(function () {
      reconnectTimer = null;
      connect();
      reconnectDelay = Math.min(reconnectDelay * 2, MAX_RECONNECT_DELAY);
    }, reconnectDelay);
  }

  function setStatus(state) {
    statusPill.className = 'status-pill status-' + state;
    statusPill.textContent = state;
  }

  sidebarToggle.addEventListener('click', function () {
    sidebar.classList.toggle('sidebar-open');
  });

  // ── Session history sidebar ──────────────────────────────
  var sessionList    = document.getElementById('session-list');
  var newSessionBtn  = document.getElementById('new-session-btn');
  var currentSessionID = null;
  var SESSIONS_STORAGE_KEY = 'claw_chat_sessions';
  var sessionMessages = {};    // sessionID → [messages HTML array]
  var CACHED_SESSION_COUNT = 50; // max cached session message lists

  // Load session list from server and render the sidebar.
  function fetchSessions() {
    var req = new XMLHttpRequest();
    req.open('GET', '/api/sessions', true);
    req.onload = function () {
      if (req.status === 200) {
        try {
          var sessions = JSON.parse(req.responseText);
          renderSessionList(sessions);
        } catch (_) { /* ignore parse errors */ }
      }
    };
    req.onerror = function () { /* server might not support sessions endpoint yet */ };
    req.send();
  }

  // Format a time string relative to now for display.
  function relativeTime(dateStr) {
    if (!dateStr) return '';
    var d = new Date(dateStr);
    var now = Date.now();
    var diff = now - d.getTime();
    if (diff < 60000) return 'Just now';
    if (diff < 3600000) return Math.floor(diff / 60000) + 'm ago';
    if (diff < 86400000) return Math.floor(diff / 3600000) + 'h ago';
    return d.toLocaleDateString();
  }

  function renderSessionList(sessions) {
    if (!sessions || !sessions.length) return;
    sessionList.innerHTML = '';
    for (var i = 0; i < sessions.length; i++) {
      var s = sessions[i];
      var li = document.createElement('li');
      li.className = 'session-item';
      if (currentSessionID === s.id) li.classList.add('active');

      var titleText = s.first_user_msg || s.last_assistant_msg || 'Session ' + (s.message_count || 0) + ' msgs';
      if (titleText.length > 50) titleText = titleText.substring(0, 47) + '…';

      var titleEl = document.createElement('span');
      titleEl.textContent = titleText;
      li.appendChild(titleEl);

      var timeEl = document.createElement('span');
      timeEl.className = 'session-time';
      timeEl.textContent = relativeTime(s.last_active_at || s.created_at);
      li.appendChild(timeEl);

      (function (sessionID) {
        li.addEventListener('click', function () {
          activateSession(sessionID);
          if (window.innerWidth < 768) sidebar.classList.remove('sidebar-open');
        });
      })(s.id);

      sessionList.appendChild(li);
    }
  }

  // Load messages for a session. If the session is currently active
  // (WS connected), we just restore from in-memory cache; otherwise we
  // show a note that only the current session's messages are available
  // in the current tab.
  function activateSession(sessionID) {
    if (sessionID === currentSessionID) return;

    // Mark active in sidebar.
    var items = sessionList.querySelectorAll('.session-item');
    for (var i = 0; i < items.length; i++) {
      items[i].classList.remove('active');
      if (items[i].dataset) items[i].classList.remove('active');
    }
    // For now, we just refresh the session list to update highlighting.
    // Full session switching (loading past conversations) requires
    // server-side session persistence which will come in a future phase.
    currentSessionID = sessionID;
    fetchSessions();
  }

  // New session: disconnect WS, clear messages, reconnect.
  function startNewSession() {
    if (ws) { ws.onclose = null; ws.close(); }
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null; }
    // Save current messages before clearing.
    if (currentSessionID) {
      sessionMessages[currentSessionID] = messagesEl.innerHTML;
      pruneSessionCache();
    }
    // Clear UI.
    messagesEl.innerHTML = '';
    currentBubble = null;
    currentBubbleContent = '';
    pendingToolCard = null;
    toolStartTime = null;
    pendingMessages = {};
    ackedMessages = {};
    currentSessionID = null;
    removeThinkingIndicator();
    hideRateLimitBanner();
    showEmptyState();
    // Reconnect will get a new session.
    setStatus('connecting');
    connect();
    // Close mobile sidebar if open.
    if (window.innerWidth < 768) sidebar.classList.remove('sidebar-open');
  }

  function pruneSessionCache() {
    var keys = Object.keys(sessionMessages);
    if (keys.length <= CACHED_SESSION_COUNT) return;
    // Remove oldest entries (by key order).
    var remove = keys.slice(0, keys.length - CACHED_SESSION_COUNT);
    for (var i = 0; i < remove.length; i++) {
      delete sessionMessages[remove[i]];
    }
  }

  if (newSessionBtn) {
    newSessionBtn.addEventListener('click', function () {
      startNewSession();
    });
  }

  // Fetch sessions on load and periodically.
  fetchSessions();
  setInterval(fetchSessions, 30000); // refresh every 30s

  setupKeyboardHandling();
  sendButton.disabled = true;

  // ── Keyboard shortcuts overlay ───────────────────────────
  var shortcutsOverlay = document.getElementById('shortcuts-overlay');
  var shortcutsToggle = document.getElementById('shortcuts-toggle');
  var shortcutsClose  = document.getElementById('shortcuts-close');
  var shortcutsBackdrop = shortcutsOverlay ? shortcutsOverlay.querySelector('.shortcuts-backdrop') : null;

  function showShortcuts() {
    if (!shortcutsOverlay) return;
    shortcutsOverlay.classList.remove('shortcuts-overlay-hidden');
    shortcutsOverlay.setAttribute('aria-hidden', 'false');
    // Focus the close button for keyboard accessibility.
    if (shortcutsClose) setTimeout(function () { shortcutsClose.focus(); }, 50);
  }

  function hideShortcuts() {
    if (!shortcutsOverlay) return;
    shortcutsOverlay.classList.add('shortcuts-overlay-hidden');
    shortcutsOverlay.setAttribute('aria-hidden', 'true');
    // Return focus to the composer.
    composerInput.focus();
  }

  function toggleShortcuts() {
    if (!shortcutsOverlay) return;
    if (shortcutsOverlay.classList.contains('shortcuts-overlay-hidden')) {
      showShortcuts();
    } else {
      hideShortcuts();
    }
  }

  // "?" button in header (touch users).
  if (shortcutsToggle) {
    shortcutsToggle.addEventListener('click', toggleShortcuts);
  }

  // Close button inside the overlay.
  if (shortcutsClose) {
    shortcutsClose.addEventListener('click', hideShortcuts);
  }

  // Click backdrop to close.
  if (shortcutsBackdrop) {
    shortcutsBackdrop.addEventListener('click', hideShortcuts);
  }

  // Keyboard listeners.
  document.addEventListener('keydown', function (e) {
    // Toggle shortcuts overlay on "?" (but not when typing in composer).
    if (e.key === '?' && !e.ctrlKey && !e.metaKey && !e.altKey && document.activeElement !== composerInput) {
      e.preventDefault();
      toggleShortcuts();
      return;
    }
    // Close shortcuts on Escape.
    if (e.key === 'Escape' && shortcutsOverlay && !shortcutsOverlay.classList.contains('shortcuts-overlay-hidden')) {
      e.preventDefault();
      hideShortcuts();
    }
  });

  connect();
})();
