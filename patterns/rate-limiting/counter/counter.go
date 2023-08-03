// ratelimit/counter/counter.go

package counter

import "time"

// Counter for rate limiting
type Counter struct {
	reqs int       // Number of current requests
	last time.Time // Time of last request
	rps  int       // Requests per second allowed
}

// New creates a new rate limiter
func New(rps int) *Counter {
	return &Counter{
		rps: rps,
	}
}

// Allow checks if request is allowed under limit
func (c *Counter) Allow() bool {
	now := time.Now()
	if c.last.IsZero() {
		// First request, do not limit
		c.last = now
		c.reqs = 1
		return true
	}

	elapsed := now.Sub(c.last)
	c.reqs++

	// Check if requests exceed RPS
	if float64(c.reqs) > float64(c.rps)*elapsed.Seconds() {
		return false
	}

	// Update last request time
	c.last = now
	c.reqs = 0
	return true
}
