/**
 * claw-code-go chat UI — client-side JavaScript
 * Mobile-first, vanilla JS. Handles WebSocket connection,
 * message rendering, and virtual keyboard detection.
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
  var currentBubble = null;   // current assistant bubble for deltas
  var currentBubbleContent = ''; // accumulated markdown content

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

  /**
   * Simple markdown-aware renderer: handles code blocks and
   * inline code. Falls back to escaped HTML for everything else.
   */
  function renderMarkdown(text) {
    // Protect against XSS by escaping HTML first
    var escaped = escapeHtml(text);

    // Replace code blocks: ```lang\n...\n```
    escaped = escaped.replace(/```(\w+)?\n([\s\S]*?)```/g, function (match, lang, code) {
      var langClass = lang ? ' class="language-' + lang + '"' : '';
      return '<pre><code' + langClass + '>' + code + '</code></pre>';
    });

    // Replace inline code: `...`
    escaped = escaped.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>');

    // Convert single newlines to <br> (preserving existing)
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
            // Group consecutive deltas into one bubble
            appendDelta(msg.text || '');
            break;
          case 'text_final':
            // Finalize the current bubble
            finalizeAssistantBubble();
            break;
          case 'tool_start':
            finalizeAssistantBubble();
            appendMessage('tool', '<strong>' + escapeHtml(msg.tool_name || 'tool') + '</strong>');
            break;
          case 'tool_done':
            if (msg.output) {
              appendMessage('tool',
                '<pre><code>' + escapeHtml(String(msg.output).substring(0, 2000)) + '</code></pre>');
            }
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
