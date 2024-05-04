package ratelimiter

import (
	"sync"

	"time"

	"golang.org/x/time/rate"
)

// Clock used to override the time used for rate limiting
type Clock interface {
	Now() time.Time
}

// RateLimiter rate limiter interface to check if we can allow a connection
type RateLimiter interface {
	Allow(clientId string) bool
}
type client struct {
	limiter *rate.Limiter
}

type ratelimit struct {
	clients map[string]*client
	burst   int
	rate    int
	mu      sync.Mutex
	clock   Clock
}

// NewRateLimiter constructor for RateLimiter
func NewRateLimiter(burst, rate int) *ratelimit {
	return &ratelimit{
		clients: make(map[string]*client),
		burst:   burst,
		rate:    rate,
		clock:   &realtime{},
	}
}

// Allow returns true if the client does not pass the rate limit
func (r *ratelimit) Allow(clientId string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.clients[clientId]
	if !ok {
		l := rate.NewLimiter(rate.Limit(r.rate), r.burst)
		c = &client{
			limiter: l,
		}
		r.clients[clientId] = c
	}
	return c.limiter.AllowN(r.clock.Now(), 1)
}

type realtime struct{}

// Now returns the now time
func (r *realtime) Now() time.Time {
	return time.Now()
}
