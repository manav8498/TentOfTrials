package gateway

import (
    "net"
    "net/http"
    "strings"
)

// ClientIP extracts the client IP from the request.
func ClientIP(r *http.Request) (string, error) {
    // Check for X-Forwarded-For header
    if xff := r.Header.Get("X-Forwarded-For"); xff!= "" {
        // X-Forwarded-For can contain multiple IPs, take the first one
        parts := strings.Split(xff, ",")
        return strings.TrimSpace(parts[0]), nil
    }

    // Check for X-Real-IP header
    if xrip := r.Header.Get("X-Real-IP"); xrip!= "" {
        return xrip, nil
    }

    // Fallback to RemoteAddr
    remoteIP, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
    if err!= nil {
        return "", err
    }
    return remoteIP, nil
}

// Example middleware using the ClientIP helper
func RateLimitMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        clientIP, err := ClientIP(r)
        if err!= nil {
            http.Error(w, "Failed to get client IP", http.StatusInternalServerError)
            return
        }
        // Use clientIP for rate limiting logic
        //...

        next.ServeHTTP(w, r)
    })
}
