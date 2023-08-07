// Package tokenbucket implements a token bucket rate limiting algorithm.
package tokenbucket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTokenBucket(t *testing.T) {

	rate := 100.0
	capacity := 1000

	tb := New(rate, capacity)

	// check rate
	if tb.Rate() != rate {
		t.Errorf("Rate does not match, expect: %f, got: %f", rate, tb.Rate())
	}

	// check capacity
	if tb.Capacity() != capacity {
		t.Errorf("Capacity does not match, expect: %d, got: %d", capacity, tb.Capacity())
	}

	// Initialize Available to 0
	if tb.Available() != 0 {
		t.Errorf("Available does not initialize to 0")
	}

	// Tokens channel created
	if tb.tokens == nil {
		t.Error("Tokens channel not initialized")
	}

	// Closed channel created
	if tb.closed == nil {
		t.Error("Closed channel not initialized")
	}

	// Started the goroutine to populate the token.
	time.Sleep(10 * time.Millisecond)
	if tb.Available() == 0 {
		t.Error("Goroutine to fill tokens not started")
	}

	// Close token bucket
	tb.Close()
}

func TestStartFillingTokens(t *testing.T) {

	rate := 100.0
	tb := New(rate, 1000)

	// Correct fill interval
	fillInterval := time.Second / time.Duration(rate)
	if fillInterval != time.Millisecond*10 {
		t.Error("Fill interval incorrect")
	}

	// Token filling goroutine started
	time.Sleep(300 * time.Millisecond)
	if tb.Available() <= 2 {
		t.Error("Goroutine not filling tokens")
	}

	// Goroutine exits after closed
	tb.Close()
	time.Sleep(200 * time.Millisecond)
	if tb.Available() > 20 {
		t.Error("Goroutine not exited after closed")
	}
}

func TestTake(t *testing.T) {

	tb := New(1000, 10)

	time.Sleep(500 * time.Millisecond)

	// Available is full before Take
	assert.Equal(t, tb.Available(), 10)

	err := tb.Take()
	assert.Nil(t, err)

	// Available reduced after Take
	assert.Equal(t, tb.Available(), 9)

	// Repeatedly Take until available is 0
	for tb.Available() > 0 {
		err = tb.Take()
		assert.Nil(t, err)
	}

	// No tokens available when Take
	err = tb.Take()
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "no tokens available")

}

func TestPut(t *testing.T) {

	tb := New(1000, 10)

	time.Sleep(500 * time.Millisecond)

	tb.Take()

	// Available is 0 before Put
	if tb.Available() != 9 {
		t.Error("available not initialized to 0")
	}

	tb.Put()

	// Available incremented after Put
	if tb.Available() != 10 {
		t.Error("available not incremented after Put")
	}

	// Repeatedly Put until full
	for tb.Available() < tb.Capacity() {
		tb.Put()
	}

	if tb.Available() != tb.Capacity() {
		t.Error("available not equals to capacity after Puts")
	}

	// Put blocks after full
	go func() {
		time.Sleep(10 * time.Millisecond)
		tb.Put()
	}()

	// Verify available not increased after 10ms
	time.Sleep(200 * time.Millisecond)
	if tb.Available() != tb.Capacity() {
		t.Error("available leaked after full")
	}

}

func TestClose(t *testing.T) {

	tb := New(1000, 10)

	// Channels are open before close
	assert.Equal(t, atomicClosedState, uint32(0))
	assert.Equal(t, atomicTokensState, uint32(0))

	// States changed after close
	tb.Close()
	assert.Equal(t, atomicClosedState, uint32(1))
	assert.Equal(t, atomicTokensState, uint32(1))

	// States stay the same after repeated close
	tb.Close()
	assert.Equal(t, atomicClosedState, uint32(1))
	assert.Equal(t, atomicTokensState, uint32(1))

	// Channels are closed
	_, closed := <-tb.closed
	assert.False(t, closed)

	_, tokens := <-tb.tokens
	assert.False(t, tokens)

}

func TestAtomicClose(t *testing.T) {

	tb := New(1000, 10)

	// Initially not closed
	assert.Equal(t, atomicClosedState, uint32(0))

	// Marked closed after call
	tb.atomicClose(tb.closed, &atomicClosedState)
	assert.Equal(t, atomicClosedState, uint32(1))

	// State stays the same after repeated calls
	tb.atomicClose(tb.closed, &atomicClosedState)
	assert.Equal(t, atomicClosedState, uint32(1))

	// Initially not closed
	assert.Equal(t, atomicTokensState, uint32(0))

	// Marked closed after call
	tb.atomicClose(tb.tokens, &atomicTokensState)
	assert.Equal(t, atomicTokensState, uint32(1))

	// State stays the same after repeated calls
	tb.atomicClose(tb.tokens, &atomicTokensState)
	assert.Equal(t, atomicTokensState, uint32(1))

}
