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
    ws.send(JSON.stringify({ type: 'user_input', text: text }));
    appendMessage('user', escapeHtml(text));
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
  function appendMessage(role, html) {
    var wrapper = document.createElement('div');
    wrapper.className = 'message message-' + role;
    wrapper.innerHTML = html;
    messagesEl.appendChild(wrapper);
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
      currentBubble = null;
      currentBubbleContent = '';
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

  function scrollToBottom() {
    requestAnimationFrame(function () {
      messagesEl.scrollTop = messagesEl.scrollHeight;
    });
  }

  // ── Connection ───────────────────────────────────────────
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
      if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null; }
    };

    ws.onmessage = function (event) {
      try {
        var msg = JSON.parse(event.data);
        switch (msg.type) {
          case 'chat_session_init': break;
          case 'text_delta':  appendDelta(msg.text || ''); break;
          case 'text_final':  finalizeAssistantBubble(); break;
          case 'tool_start':  startToolCard(msg.tool_name, msg.input_summary || ''); break;
          case 'tool_done':   finishToolCard(msg.output || ''); break;
          case 'permission_ask': showPermissionDialog(msg); break;
          case 'error':
            finalizeAssistantBubble();
            appendMessage('error', 'Error: ' + escapeHtml(msg.error || 'unknown'));
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

  setupKeyboardHandling();
  sendButton.disabled = true;
  connect();
})();
