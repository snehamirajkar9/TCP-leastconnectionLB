package balancer

import (
	"net"

	"github.com/snehamirajkar9/TCP-leastconnectionLB/balancer/forwarder"
	"github.com/snehamirajkar9/TCP-leastconnectionLB/balancer/ratelimiter"
)

// Balancer provides load balancing functionality
type Balancer interface {
	Balance(conn net.Conn, clientId string, allowedUpstreams []string) error
	Stop()
}

// Config holds balancer configuration
type Config struct {
	Burst     int
	Rate      int
	Upstreams []string
}

type balance struct {
	forwarder   forwarder.Forwarder
	rateLimiter ratelimiter.RateLimiter
}

// NewBalancer constructor for Balancer
func NewBalancer(c Config) *balance {
	b := balance{
		forwarder:   forwarder.NewForwarder(c.Upstreams),
		rateLimiter: ratelimiter.NewRateLimiter(c.Burst, c.Rate),
	}
	return &b
}

// Balance load balance a connect by calling RateLimier with the clientId and the forwarder with
// the list of allowed upstreams
func (b *balance) Balance(conn net.Conn, clientId string, allowedUpstreams []string) error {
	var err error
	if b.rateLimiter.Allow(clientId) {
		err = b.forwarder.Forward(conn, allowedUpstreams)
	} else {
		err = &RateLimiterError{}
	}
	return err
}

// RateLimiterError error used when connection has been rate limiter
type RateLimiterError struct{}

// Error string version of the error
func (e *RateLimiterError) Error() string {
	return "client rate limited"
}

func (b *balance) Stop() {
	b.forwarder.Stop()
}
