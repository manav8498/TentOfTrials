package gateway

import (
    "net"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
)

func TestClientIP(t *testing.T) {
    tests := []struct {
        name     string
        headers  map[string]string
        remote   string
        expected string
    }{
        {"Direct client with host:port", nil, "192.168.1.1:8080", "192.168.1.1"},
        {"Direct client with bare IP", nil, "192.168.1.1", "192.168.1.1"},
        {"Malformed RemoteAddr", nil, "invalid:8080", ""},
        {"X-Forwarded-For header", map[string]string{"X-Forwarded-For": "10.0.0.1, 10.0.0.2"}, "192.168.1.1:8080", "10.0.0.1"},
        {"X-Real-IP header", map[string]string{"X-Real-IP": "10.0.0.1"}, "192.168.1.1:8080", "10.0.0.1"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r, _ := http.NewRequest("GET", "/", nil)
            r.RemoteAddr = tt.remote
            for k, v := range tt.headers {
                r.Header.Set(k, v)
            }

            ip, err := ClientIP(r)
            if err!= nil && tt.expected!= "" {
                t.Errorf("ClientIP() error = %v, want no error", err)
                return
            }
            if ip!= tt.expected {
                t.Errorf("ClientIP() = %v, want %v", ip, tt.expected)
            }
        })
    }
}

func TestRateLimitMiddleware(t *testing.T) {
    ts := httptest.NewServer(RateLimitMiddleware(http.HandlerFunc(HomeHandler)))
    defer ts.Close()

    res, err := http.Get(ts.URL)
    if err!= nil {
        t.Fatalf("Failed to make request: %v", err)
    }
    if res.StatusCode!= http.StatusOK {
        t.Errorf("Expected status OK, got %d", res.StatusCode)
    }
}

func TestLogRequestHandler(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(LogRequestHandler))
    defer ts.Close()

    res, err := http.Get(ts.URL + "/log")
    if err!= nil {
        t.Fatalf("Failed to make request: %v", err)
    }
    if res.StatusCode!= http.StatusOK {
        t.Errorf("Expected status OK, got %d", res.StatusCode)
    }
}
