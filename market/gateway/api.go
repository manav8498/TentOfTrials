package gateway

import (
    "log"
    "net/http"
)

// Example API handler using the ClientIP helper
func LogRequestHandler(w http.ResponseWriter, r *http.Request) {
    clientIP, err := ClientIP(r)
    if err!= nil {
        http.Error(w, "Failed to get client IP", http.StatusInternalServerError)
        return
    }
    log.Printf("Incoming request from %s", clientIP)
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Request logged"))
}

// Setup routes and middleware
func SetupRoutes() *http.ServeMux {
    mux := http.NewServeMux()
    mux.HandleFunc("/log", LogRequestHandler)
    mux.Handle("/", RateLimitMiddleware(http.HandlerFunc(HomeHandler)))
    return mux
}

// HomeHandler is a simple home handler
func HomeHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Welcome to the Gateway"))
}
