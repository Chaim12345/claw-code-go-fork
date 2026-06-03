/**
 * ESLint flat config for claw-code-go web assets.
 * Targets browser JavaScript (ES5-compatible, no ES6 modules).
 */

module.exports = [
  {
    languageOptions: {
      ecmaVersion: 5,
      sourceType: "script",
      globals: {
        // Browser globals
        window: "readonly",
        document: "readonly",
        console: "readonly",
        localStorage: "readonly",
        sessionStorage: "readonly",
        navigator: "readonly",
        fetch: "readonly",
        Request: "readonly",
        Response: "readonly",
        URL: "readonly",
        URLSearchParams: "readonly",
        WebSocket: "readonly",
        XMLHttpRequest: "readonly",
        Blob: "readonly",
        FileReader: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
        setInterval: "readonly",
        clearInterval: "readonly",
        requestAnimationFrame: "readonly",
        crypto: "readonly",
        self: "readonly",
        caches: "readonly",
        clients: "readonly",
        getComputedStyle: "readonly",
        // ES5 Array/Object/JSON are built-in
      },
    },
    rules: {
      // Errors (things we should fix)
      "no-undef": "error",
      "no-unused-vars": ["warn", { args: "none", caughtErrors: "none" }],
      "no-redeclare": "error",
      "no-dupe-keys": "error",
      "no-duplicate-case": "error",
      "no-empty": "warn",
      "no-extra-semi": "warn",
      "no-func-assign": "error",
      "no-constant-condition": "warn",
      // Style (warnings only, our JS is ES5-style)
      "semi": ["warn", "always"],
      "eqeqeq": ["warn", "smart"],
    },
  },
  {
    // Service worker has its own globals
    files: ["internal/web/static/sw.js"],
    languageOptions: {
      ecmaVersion: 2017,
      sourceType: "script",
      globals: {
        self: "readonly",
        caches: "readonly",
        clients: "readonly",
        fetch: "readonly",
        Request: "readonly",
        Response: "readonly",
        URL: "readonly",
        Promise: "readonly",
      },
    },
  },
  {
    // Ignore third-party code
    ignores: [
      "internal/web/static/js/@wterm/**",
      "node_modules/**",
    ],
  },
];