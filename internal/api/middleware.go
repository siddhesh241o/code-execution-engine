package api

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	clients = make(map[string]*client)
	mu      sync.Mutex
)

func init() {
	go cleanupClients()
}

func cleanupClients() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, c := range clients {
			if time.Since(c.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		mu.Unlock()
	}
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(r.URL.Path == "/api/execute") {
			next.ServeHTTP(w, r)
			return
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Every(time.Minute/5), 10)}
		}
		clients[ip].lastSeen = time.Now()
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func RequireFrontendAuth(next http.Handler, expectedSecret string, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !(r.URL.Path == "/api/execute") {
			next.ServeHTTP(w, r)
			return
		}

		if expectedSecret == "" {
			http.Error(w, "frontend auth is not configured", http.StatusServiceUnavailable)
			return
		}

		if allowedOrigin != "" {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" && origin != allowedOrigin {
				http.Error(w, "forbidden origin", http.StatusForbidden)
				return
			}
		}

		if r.Header.Get("X-Frontend-Secret") != expectedSecret {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
