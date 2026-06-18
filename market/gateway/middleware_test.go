package gateway

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		// Direct clients
		{
			name:       "host:port RemoteAddr",
			remoteAddr: "192.168.1.100:54321",
			want:       "192.168.1.100",
		},
		{
			name:       "bare IP RemoteAddr",
			remoteAddr: "10.0.0.1",
			want:       "10.0.0.1",
		},
		{
			name:       "IPv6 RemoteAddr",
			remoteAddr: "[::1]:54321",
			want:       "::1",
		},
		// Malformed RemoteAddr fallback
		{
			name:       "malformed RemoteAddr — no port",
			remoteAddr: "not-an-address",
			want:       "not-an-address",
		},
		{
			name:       "empty RemoteAddr",
			remoteAddr: "",
			want:       "",
		},
		// X-Forwarded-For header
		{
			name:       "X-Forwarded-For single IP",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.5"},
			want:       "203.0.113.5",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.5, 198.51.100.1, 192.0.2.1"},
			want:       "203.0.113.5",
		},
		{
			name:       "X-Forwarded-For with spaces",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Forwarded-For": "  203.0.113.5  , 198.51.100.1"},
			want:       "203.0.113.5",
		},
		// X-Real-IP header
		{
			name:       "X-Real-IP takes precedence over RemoteAddr",
			remoteAddr: "10.0.0.1:12345",
			headers:    map[string]string{"X-Real-IP": "203.0.113.10"},
			want:       "203.0.113.10",
		},
		// X-Forwarded-For takes precedence over X-Real-IP
		{
			name:       "X-Forwarded-For overrides X-Real-IP",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.5",
				"X-Real-IP":       "198.51.100.1",
			},
			want: "203.0.113.5",
		},
		// X-Forwarded-For takes precedence over X-Real-IP
		{
			name:       "X-Real-IP with empty X-Forwarded-For",
			remoteAddr: "10.0.0.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "",
				"X-Real-IP":       "203.0.113.99",
			},
			want: "203.0.113.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			got := getClientIP(req)
			if got != tt.want {
				t.Errorf("getClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetClientIP_IntegrationWithMiddleware(t *testing.T) {
	// Verify that both logging and rate-limit middleware use getClientIP
	// and resolve consistent client IP values.
	t.Run("consistent IP in LoggingMiddleware", func(t *testing.T) {
		var capturedIP string
		handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedIP = getClientIP(r)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.50:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.42")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if capturedIP != "203.0.113.42" {
			t.Errorf("LoggingMiddleware resolved IP = %q, want %q", capturedIP, "203.0.113.42")
		}
	})

	t.Run("consistent IP in RateLimitMiddleware", func(t *testing.T) {
		var capturedIP string
		rateLimiter := RateLimitMiddleware(10.0, 5)
		handler := rateLimiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedIP = getClientIP(r)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.50:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.42")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if capturedIP != "203.0.113.42" {
			t.Errorf("RateLimitMiddleware resolved IP = %q, want %q", capturedIP, "203.0.113.42")
		}
	})
}
