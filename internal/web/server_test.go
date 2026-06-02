package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_IndexServesHTML(t *testing.T) {
	s := NewServer(Config{Addr: "127.0.0.1:0"})
	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	body := string(buf[:n])
	if !strings.Contains(body, "claw-code-go") {
		t.Errorf("body missing title; got: %q", body)
	}
}

func TestServer_Healthz(t *testing.T) {
	s := NewServer(Config{Addr: "127.0.0.1:0"})
	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestServer_StaticAssetsServed(t *testing.T) {
	s := NewServer(Config{Addr: "127.0.0.1:0"})
	ts := httptest.NewServer(s.routes())
	defer ts.Close()

	// Pick a few representative assets.
	for _, p := range []string{
		"/static/wasm/wterm.wasm",
		"/static/css/terminal.css",
		"/static/js/@wterm/dom/wterm.js",
		"/static/js/@wterm/dom/index.js",
		"/static/js/@wterm/core/transport.js",
		"/static/js/@wterm/core/index.js",
	} {
		resp, err := http.Get(ts.URL + p)
		if err != nil {
			t.Errorf("%s: %v", p, err)
			continue
		}
		if resp.StatusCode != 200 {
			t.Errorf("%s: status %d", p, resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func TestParseControlMessage(t *testing.T) {
	// sanity: handleControlMessage should not write to PTY.
	sess := &ptySession{}
	handled, err := sess.handleControlMessage([]byte(`{"type":"resize","cols":80,"rows":24}`))
	if !handled {
		t.Error("expected handled=true for resize")
	}
	// We can't actually call Setsize in a unit test (no PTY), but
	// the call path should be exercised. The error returned is fine.
	_ = err

	handled, err = sess.handleControlMessage([]byte(`not json`))
	if !handled {
		t.Error("expected handled=true for malformed json (to suppress passthrough)")
	}
	if err == nil {
		t.Error("expected error for malformed json")
	}

	handled, _ = sess.handleControlMessage([]byte(`{"type":"unknown"}`))
	if !handled {
		t.Error("expected handled=true for unknown control type")
	}
}
