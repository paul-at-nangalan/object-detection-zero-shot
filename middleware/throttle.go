package middleware

import (
	"net/http"
	"sync"
	"time"
)

type ThrottleMiddleware struct {
	mu            sync.RWMutex
	requests      map[string][]time.Time
	maxRequests   int
	periodInHours int
}

// NewThrottleMiddleware creates a new throttle middleware with the specified request limit per period
func NewThrottleMiddleware(maxRequests, periodInHours int) *ThrottleMiddleware {
	return &ThrottleMiddleware{
		requests:      make(map[string][]time.Time),
		maxRequests:   maxRequests,
		periodInHours: periodInHours,
	}
}

// Wrap wraps an http.HandlerFunc with the throttling middleware
func (t *ThrottleMiddleware) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the real IP from Cloudflare headers
		ip := r.Header.Get("CF-Connecting-IP")
		if ip == "" {
			ip = r.RemoteAddr // Fallback to remote address if CF header not present
		}
		if !t.allowRequest(ip) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// allowRequest checks if a request from the given IP should be allowed
func (t *ThrottleMiddleware) allowRequest(ip string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-time.Duration(t.periodInHours) * time.Hour)
	// Clean up old requests
	times, exists := t.requests[ip]
	if exists {
		var validTimes []time.Time
		for _, reqTime := range times {
			if reqTime.After(cutoff) {
				validTimes = append(validTimes, reqTime)
			}
		}
		t.requests[ip] = validTimes
	} else {
		t.requests[ip] = make([]time.Time, 0)
	}
	// Check if rate limit is exceeded
	if len(t.requests[ip]) >= t.maxRequests {
		return false
	}
	// Add new request
	t.requests[ip] = append(t.requests[ip], now)
	return true
}
