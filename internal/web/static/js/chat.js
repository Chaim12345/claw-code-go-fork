/**
 * claw-code-go chat UI — client-side JavaScript
 * Mobile-first, vanilla JS. Handles WebSocket connection,
 * message rendering, and virtual keyboard detection.
 */

(function () {
  'use strict';

  // ── DOM refs ──────────────────────────────────────────────
  const composer     = document.getElementById('composer');
  const composerInput = document.getElementById('composer-input');
  const sendButton   = document.getElementById('send-button');
  const messagesEl   = document.getElementById('messages');
  const statusPill   = document.getElementById('connection-status');
  const sidebarToggle = document.getElementById('sidebar-toggle');
  const sidebar       = document.getElementById('sidebar');

  // ── State ─────────────────────────────────────────────────
  let ws = null;
  let reconnectTimer = null;
  let reconnectDelay = 1000;
  const MAX_RECONNECT_DELAY = 30000;

  // ── Visual Viewport (virtual keyboard) handling ───────────
  /**
   * On mobile, the visualViewport API fires 'resize' when the
   * virtual keyboard appears or disappears. We adjust the
   * composer position and scroll messages to the bottom so
   * the user can see the latest content above the keyboard.
   */
  function handleViewportResize() {
    if (!window.visualViewport) return;

    var viewport = window.visualViewport;
    var keyboardHeight = window.innerHeight - viewport.height;

    // Adjust the composer bottom padding so it sits above
    // the virtual keyboard. The --keyboard-offset custom
    // property can be read by CSS if needed, but here we
    // directly scroll to ensure the last message is visible.
    if (keyboardHeight > 100) {
      // Keyboard is likely visible
      composer.style.paddingBottom = (keyboardHeight + 8) + 'px';
      scrollToBottom();
    } else {
      // Keyboard hidden — reset to safe-area default
      composer.style.paddingBottom = '';
    }

    // If the viewport is offset (iOS Safari scrolls the page
    // when the keyboard opens), we compensate by adjusting
    // the messages scroll position.
    if (viewport.offsetTop > 0) {
      messagesEl.scrollTop += viewport.offsetTop;
    }
  }

  /**
   * Sets up virtual keyboard detection via visualViewport API.
   * This is the primary mechanism for handling mobile keyboards.
   * Falls back gracefully on desktop where visualViewport is
   * either absent or matches window.innerHeight exactly.
   */
  function setupKeyboardHandling() {
    if (window.visualViewport) {
      window.visualViewport.addEventListener('resize', handleViewportResize);
      window.visualViewport.addEventListener('scroll', handleViewportResize);
    }
  }

  // ── Composer auto-grow ────────────────────────────────────
  composerInput.addEventListener('input', function () {
    // Auto-grow the textarea up to 8 rows
    this.style.height = 'auto';
    var lineHeight = parseFloat(getComputedStyle(this).lineHeight) || 24;
    var maxHeight = lineHeight * 8;
    this.style.height = Math.min(this.scrollHeight, maxHeight) + 'px';

    // Enable/disable send button based on content
    sendButton.disabled = this.value.trim().length === 0;

    // Adjust viewport on mobile after resize
    if (window.visualViewport) {
      handleViewportResize();
    }
  });

  // ── Form submit ───────────────────────────────────────────
  composer.addEventListener('submit', function (e) {
    e.preventDefault();
    var text = composerInput.value.trim();
    if (!text || !ws || ws.readyState !== WebSocket.OPEN) return;

    // Send user input over WebSocket
    ws.send(JSON.stringify({ type: 'user_input', text: text }));

    // Render user message
    appendMessage('user', text);

    // Clear input and reset height
    composerInput.value = '';
    composerInput.style.height = '';
    sendButton.disabled = true;
    composerInput.focus();
    scrollToBottom();
  });

  // ── Message rendering ─────────────────────────────────────
  function appendMessage(role, text) {
    var div = document.createElement('div');
    div.className = 'message message-' + role;
    div.textContent = text;
    messagesEl.appendChild(div);
    scrollToBottom();
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

    // Include auth token from URL query param if present
    var params = new URLSearchParams(window.location.search);
    var token = params.get('token');
    if (token) {
      wsURL += '?token=' + encodeURIComponent(token);
    }

    ws = new WebSocket(wsURL);

    ws.onopen = function () {
      setStatus('connected');
      reconnectDelay = 1000;
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
    };

    ws.onmessage = function (event) {
      try {
        var msg = JSON.parse(event.data);
        switch (msg.type) {
          case 'chat_session_init':
            // Session started — could show a welcome message
            break;
          case 'text_delta':
            // For now, append as simple text. Message grouping
            // and rich rendering will be added in later items.
            appendMessage('assistant', msg.text || '');
            break;
          case 'text_final':
            // Final text from a turn
            break;
          case 'tool_start':
            appendMessage('tool', '[tool] ' + (msg.tool_name || 'unknown'));
            break;
          case 'error':
            appendMessage('error', 'Error: ' + (msg.error || 'unknown'));
            break;
          case 'done':
            break;
          default:
            break;
        }
      } catch (err) {
        // Non-JSON message — ignore
      }
    };

    ws.onclose = function () {
      setStatus('disconnected');
      scheduleReconnect();
    };

    ws.onerror = function () {
      // onclose will fire after this
    };
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

  // ── Sidebar toggle (mobile hamburger) ─────────────────────
  sidebarToggle.addEventListener('click', function () {
    sidebar.classList.toggle('sidebar-open');
  });

  // ── Init ──────────────────────────────────────────────────
  setupKeyboardHandling();
  sendButton.disabled = true;
  connect();
})();
