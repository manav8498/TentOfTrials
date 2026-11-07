package gateway

import (
	"net/http"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expected   string
	}{
		{
			name:       "Direct client host:port",
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "Bare IP RemoteAddr",
			remoteAddr: "10.0.0.1",
			expected:   "10.0.0.1",
		},
		{
			name:       "Malformed RemoteAddr fallback",
			remoteAddr: "malformed-addr-no-port-or-ip",
			expected:   "malformed-addr-no-port-or-ip",
		},
		{
			name:       "X-Forwarded-For simple",
			remoteAddr: "127.0.0.1:8080",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
			},
			expected: "203.0.113.195",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			remoteAddr: "127.0.0.1:8080",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195, 70.41.3.18, 150.172.238.178",
			},
			expected: "203.0.113.195",
		},
		{
			name:       "X-Real-IP",
			remoteAddr: "127.0.0.1:8080",
			headers: map[string]string{
				"X-Real-IP": "198.51.100.2",
			},
			expected: "198.51.100.2",
		},
		{
			name:       "X-Forwarded-For overrides X-Real-IP",
			remoteAddr: "127.0.0.1:8080",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
				"X-Real-IP":       "198.51.100.2",
			},
			expected: "203.0.113.195",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remoteAddr
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			ip := getClientIP(req)
			if ip != tc.expected {
				t.Errorf("Expected IP %q, got %q", tc.expected, ip)
			}
		})
	}
}
