package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter is a fixed-window counter per key (client IP).
// ponytail: in-memory, so each instance counts separately; move counts to Postgres or
// Redis if the API runs on more than a couple of instances.
type RateLimiter struct {
	mu        sync.Mutex
	limit     int
	window    time.Duration
	hits      map[string]*window
	lastSweep time.Time
}

type window struct {
	count int
	reset time.Time
}

func NewRateLimiter(limit int, per time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, window: per, hits: map[string]*window{}, lastSweep: time.Now()}
}

// Allow records a hit for key and reports whether it is within the limit,
// plus how long until the key's window resets.
func (l *RateLimiter) Allow(key string) (bool, time.Duration) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if now.Sub(l.lastSweep) > l.window { // keep memory bounded
		for k, w := range l.hits {
			if now.After(w.reset) {
				delete(l.hits, k)
			}
		}
		l.lastSweep = now
	}

	w, ok := l.hits[key]
	if !ok || now.After(w.reset) {
		w = &window{reset: now.Add(l.window)}
		l.hits[key] = w
	}
	w.count++
	return w.count <= l.limit, w.reset.Sub(now)
}

// Middleware rejects clients over the limit with 429. Keyed on c.ClientIP(), so set
// the router's TrustedPlatform when running behind a proxy.
func (l *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if ok, retry := l.Allow(c.ClientIP()); !ok {
			c.Header("Retry-After", strconv.Itoa(int(retry.Seconds())+1))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many attempts. Please wait a few minutes and try again."})
			return
		}
		c.Next()
	}
}
