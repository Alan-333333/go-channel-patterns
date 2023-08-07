// Package tokenbucket implements a token bucket rate limiting algorithm.
package tokenbucket

import (
	"errors"
	"sync/atomic"
	"time"
)

// TokenBucket implements a token bucket that fills tokens at the specified rate.
// It allows limiting access to resources by rate.
type TokenBucket struct {

	// Rate tokens are added to the bucket per second (REQs/sec)
	rate float64

	// Capacity is the maximum number of tokens the bucket can hold
	capacity int

	// Available tokens that can be taken
	available int

	// Channel used to receive and return tokens
	tokens chan struct{}

	// Channel signaled when bucket is closed
	closed chan struct{}
}

// atomicClosedState and atomicTokensState are used to save the closed state of each channel
var atomicClosedState uint32
var atomicTokensState uint32

// New creates a new token bucket with the given rate and capacity.
func New(rate float64, capacity int) *TokenBucket {

	tb := &TokenBucket{
		rate:      rate,
		capacity:  capacity,
		available: 0,
		tokens:    make(chan struct{}, capacity),
		closed:    make(chan struct{}),
	}

	// Start goroutine to fill tokens
	go startFillingTokens(tb, rate)

	return tb
}

// startFillingTokens fills tokens at the rate
func startFillingTokens(tb *TokenBucket, rate float64) {

	fillInterval := time.Second / time.Duration(rate)

	for {
		select {
		case <-time.After(fillInterval):
			tb.fillToken()
		case <-tb.closed:
			return
		}
	}
}

// fillToken adds a token if available tokens is less than capacity.
func (tb *TokenBucket) fillToken() {

	if tb.available < tb.capacity {
		select {
		case tb.tokens <- struct{}{}: // Add new token
			tb.available++
		default: // Bucket full, do nothing
		}
	}
}

// Take retrieves a token from the bucket. It blocks if no tokens available.
func (tb *TokenBucket) Take() error {
	if tb.available <= 0 {
		return errors.New("no tokens available")
	}

	<-tb.tokens
	tb.available--

	return nil
}

// Put returns a token back to the bucket.
func (tb *TokenBucket) Put() error {

	// Checks if the current available value exceeds capacity.
	if tb.available >= tb.capacity {
		return errors.New("available exceeds capacity")
	}

	// Waiting for token slot
	select {
	case <-tb.tokens:

	// Check that the channel is closed
	case _, ok := <-tb.closed:
		if !ok {
			return errors.New("token bucket closed")
		}
	}

	// Trying to send a token
	select {
	case tb.tokens <- struct{}{}:

	// Check if the send was successful
	default:
		return errors.New("fail to send token")
	}

	// add available
	tb.available++

	return nil
}

// Rate returns the fill rate of the bucket.
func (tb *TokenBucket) Rate() float64 {
	return tb.rate
}

// Capacity returns the capacity of the bucket.
func (tb *TokenBucket) Capacity() int {
	return tb.capacity
}

// Available returns the number of available tokens.
func (tb *TokenBucket) Available() int {
	return tb.available
}

// Wait blocks until a token becomes available.
func (tb *TokenBucket) Wait() {
	<-tb.tokens
}

// Close stops the filling goroutine and closes channels.
func (tb *TokenBucket) Close() {
	// Close closed channel
	tb.atomicClose(tb.closed, &atomicClosedState)
	// Close Token channel
	tb.atomicClose(tb.tokens, &atomicTokensState)

	tb.available = 0
}

// atomicClose atomically closes the given channel
// No operation is performed if it is already closed
// The state parameter is used to save the closed state of the channel
func (tb *TokenBucket) atomicClose(ch chan struct{}, state *uint32) {
	// CAS operation to close the channel
	// Do not perform a close if the state is already closed.
	if atomic.CompareAndSwapUint32(state, 0, 1) {
		close(ch)
	}
}
