/**
 * claw-code-go chat UI — client-side JavaScript
 * Mobile-first, vanilla JS. Handles WebSocket connection,
 * message rendering, virtual keyboard, tool-call cards.
 */

(function () {
  'use strict';

  // ── DOM refs ──────────────────────────────────────────────
  var composer       = document.getElementById('composer');
  var composerInput  = document.getElementById('composer-input');
  var sendButton     = document.getElementById('send-button');
  var messagesEl     = document.getElementById('messages');
  var statusPill     = document.getElementById('connection-status');
  var sidebarToggle  = document.getElementById('sidebar-toggle');
  var sidebar        = document.getElementById('sidebar');

  // ── State ─────────────────────────────────────────────────
  var ws = null;
  var reconnectTimer = null;
  var reconnectDelay = 1000;
  var MAX_RECONNECT_DELAY = 30000;
  var currentBubble = null;
  var currentBubbleContent = '';
  var toolStartTime = null;     // timestamp when tool_start fired
  var pendingToolCard = null;   // DOM element for current tool card

  // ── Visual Viewport (virtual keyboard) handling ───────────
  function handleViewportResize() {
    if (!window.visualViewport) return;
    var viewport = window.visualViewport;
    var keyboardHeight = window.innerHeight - viewport.height;
    if (keyboardHeight > 100) {
      composer.style.paddingBottom = (keyboardHeight + 8) + 'px';
      scrollToBottom();
    } else {
      composer.style.paddingBottom = '';
    }
    if (viewport.offsetTop > 0) {
      messagesEl.scrollTop += viewport.offsetTop;
    }
  }

  function setupKeyboardHandling() {
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', handleViewportResize);
      window.visualViewport.addEventListener('scroll', handleViewportResize);
    }
  }

  // ── Composer auto-grow ────────────────────────────────────
  composerInput.addEventListener('input', function () {
    this.style.height = 'auto';
    var lineHeight = parseFloat(getComputedStyle(this).lineHeight) || 24;
    var maxHeight = lineHeight * 8;
    this.style.height = Math.min(this.scrollHeight, maxHeight) + 'px';
    sendButton.disabled = this.value.trim().length === 0;
    if (window.visualViewport) handleViewportResize();
  });

  // ── Form submit ───────────────────────────────────────────
  composer.addEventListener('submit', function (e) {
    e.preventDefault();
    var text = composerInput.value.trim();
    if (!text || !ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ type: 'user_input', text: text }));
    appendMessage('user', escapeHtml(text));
    composerInput.value = '';
    composerInput.style.height = '';
    sendButton.disabled = true;
    composerInput.focus();
    scrollToBottom();
  });

  // ── Message rendering with delta grouping ─────────────────
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

  // ── Tool-call cards ───────────────────────────────────────
  function createToolCard(toolName, inputSummary) {
    var card = document.createElement('div');
    card.className = 'message message-tool';

    var summary = document.createElement('div');
    summary.className = 'tool-summary';

    var icon = document.createElement('span');
    icon.className = 'tool-icon';
    icon.textContent = '\u2699'; // gear icon
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

    // Collapsible details block
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

    // Copy button
    var copyBtn = document.createElement('button');
    copyBtn.className = 'tool-copy-btn';
    copyBtn.textContent = 'Copy';
    copyBtn.title = 'Copy output to clipboard';
    copyBtn.addEventListener('click', function () {
      navigator.clipboard.writeText(code.textContent).then(function () {
        copyBtn.textContent = 'Copied!';
        setTimeout(function () { copyBtn.textContent = 'Copy'; }, 2000);
      }).catch(function () {
        copyBtn.textContent = 'Error';
      });
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

    // Update timer every 100ms
    var updateTimer = function () {
      if (pendingToolCard !== result) return;
      var elapsed = ((Date.now() - toolStartTime) / 1000).toFixed(1);
      result.timer.textContent = elapsed + 's';
      if (pendingToolCard === result) {
        requestAnimationFrame(function () {
          setTimeout(updateTimer, 100);
        });
      }
    };
    updateTimer();
  }

  function finishToolCard(output) {
    if (pendingToolCard) {
      pendingToolCard.code.textContent = String(output || '').substring(0, 2000);
      // Stop timer updates by clearing pendingToolCard
      pendingToolCard = null;
      toolStartTime = null;
    }
  }

  // ── Markdown renderer ─────────────────────────────────────
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

  // ── Connection management ─────────────────────────────────
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
          case 'chat_session_init':
            break;
          case 'text_delta':
            appendDelta(msg.text || '');
            break;
          case 'text_final':
            finalizeAssistantBubble();
            break;
          case 'tool_start':
            startToolCard(msg.tool_name, msg.input_summary || '');
            break;
          case 'tool_done':
            finishToolCard(msg.output || '');
            break;
          case 'error':
            finalizeAssistantBubble();
            appendMessage('error', 'Error: ' + escapeHtml(msg.error || 'unknown'));
            break;
          case 'done':
            finalizeAssistantBubble();
            break;
          default:
            break;
        }
      } catch (err) {
        // ignore non-JSON
      }
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

  // ── Sidebar toggle ────────────────────────────────────────
  sidebarToggle.addEventListener('click', function () {
    sidebar.classList.toggle('sidebar-open');
  });

  // ── Init ──────────────────────────────────────────────────
  setupKeyboardHandling();
  sendButton.disabled = true;
  connect();
})();
