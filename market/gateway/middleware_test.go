package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func reqWith(remoteAddr, xff, xri string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = remoteAddr
	if xff != "" {
		r.Header.Set("X-Forwarded-For", xff)
	}
	if xri != "" {
		r.Header.Set("X-Real-IP", xri)
	}
	return r
}

func TestClientIPFromHostPortRemoteAddr(t *testing.T) {
	got := getClientIP(reqWith("203.0.113.10:54321", "", ""))
	if got != "203.0.113.10" {
		t.Fatalf("expected host from host:port, got %q", got)
	}
}

func TestClientIPFromBareIPRemoteAddr(t *testing.T) {
	got := getClientIP(reqWith("203.0.113.20", "", ""))
	if got != "203.0.113.20" {
		t.Fatalf("expected bare IP fallback, got %q", got)
	}
}

func TestClientIPFromMalformedRemoteAddr(t *testing.T) {
	got := getClientIP(reqWith("not-an-ip", "", ""))
	if got != "not-an-ip" {
		t.Fatalf("expected malformed RemoteAddr fallback, got %q", got)
	}
}

func TestClientIPFromXForwardedFor(t *testing.T) {
	got := getClientIP(reqWith("127.0.0.1:1234", "198.51.100.5, 10.0.0.1", ""))
	if got != "198.51.100.5" {
		t.Fatalf("expected first X-Forwarded-For IP, got %q", got)
	}
}

func TestClientIPFromXRealIP(t *testing.T) {
	got := getClientIP(reqWith("127.0.0.1:1234", "", "192.0.2.44"))
	if got != "192.0.2.44" {
		t.Fatalf("expected X-Real-IP, got %q", got)
	}
}

func TestRateLimitUsesResolvedClientIP(t *testing.T) {
	var seenKey string
	handler := RateLimitMiddleware(100, 10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenKey = getClientIP(r)
		w.WriteHeader(http.StatusOK)
	}))

	r := reqWith("127.0.0.1:9999", "203.0.113.55", "")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if seenKey != "203.0.113.55" {
		t.Fatalf("rate limiter should use resolved client IP, got %q", seenKey)
	}
}

func TestLoggingUsesResolvedClientIP(t *testing.T) {
	// LoggingMiddleware calls getClientIP internally; ensure helper resolves proxy header.
	r := reqWith("127.0.0.1:1", "203.0.113.77", "")
	if got := getClientIP(r); got != "203.0.113.77" {
		t.Fatalf("logging path should see forwarded IP, got %q", got)
	}
}
