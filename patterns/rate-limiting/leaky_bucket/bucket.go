// ratelimit/leaky_bucket/bucket.go

package leakybucket

import (
	"time"
)

// LeakyBucket rate limiter
type LeakyBucket struct {
	capacity int     // Bucket capacity
	rate     float64 // Outflow rate (REQs/sec)

	requests int       // Current number of requests
	lastTime time.Time // Time of last request
}

// New creates a leaky bucket limiter
func New(capacity, rate int) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		rate:     float64(rate),
	}
}

// Allow checks if a request should be limited
func (b *LeakyBucket) Allow() bool {
	now := time.Now()
	b.requests++

	if b.lastTime.IsZero() {
		// First request, allow
		b.requests = 0
		b.lastTime = now
		return true
	}

	if b.requests >= b.capacity {
		// Not enought capacity, limit
		return false
	}

	// Calculate outflow for this request
	elapsed := now.Sub(b.lastTime).Seconds()
	outflow := elapsed * b.rate

	// Allow if outflow >= requests
	if int(outflow) >= b.requests {
		b.requests = 0
		b.lastTime = now
		return true
	}
	// Not enough outflow, limit
	return false
}
