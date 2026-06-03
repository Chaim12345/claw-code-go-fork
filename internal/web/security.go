package web

import "net/http"

// securityHeadersMiddleware adds HTTP security headers to every
// response. The Content-Security-Policy includes 'wasm-unsafe-eval'
// because the wterm terminal emulator uses WebAssembly and needs
// eval-like capabilities for its WASM runtime.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'wasm-unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"connect-src 'self' ws: wss:; "+
				"img-src 'self' data:; "+
				"font-src 'self'; "+
				"manifest-src 'self'; "+
				"frame-ancestors 'none'; "+
				"base-uri 'self'; "+
				"form-action 'self'")

		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Permissions-Policy",
			"camera=(), microphone=(), geolocation=()")

		next.ServeHTTP(w, r)
	})
}