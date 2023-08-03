// ratelimit/counter/counter_test.go

package counter

import (
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("check fields are set", func(t *testing.T) {
		rps := 10
		limiter := New(rps)

		if limiter.rps != rps {
			t.Errorf("rps field not set correctly, expected=%d, got=%d", rps, limiter.rps)
		}

		if !limiter.last.IsZero() {
			t.Error("last field should be zero for new limiter")
		}

		if limiter.reqs != 0 {
			t.Error("reqs field should be zero for new limiter")
		}
	})

	t.Run("zero rps", func(t *testing.T) {
		limiter := New(0)
		if limiter.rps != 0 {
			t.Error("rps should be allowed to be set to zero")
		}
	})
}
func TestAllow(t *testing.T) {

	t.Run("first request", func(t *testing.T) {
		limiter := New(10)

		if !limiter.Allow() {
			t.Error("First request should always be allowed")
		}

		if limiter.last.IsZero() {
			t.Error("Last should be updated after first request")
		}
	})

	t.Run("under limit", func(t *testing.T) {
		for i := 0; i < 9; i++ {
			limiter := New(10)
			if !limiter.Allow() {
				t.Error("Request under limit should be allowed")
			}
		}
	})

	t.Run("over limit", func(t *testing.T) {
		limiter := New(1)
		limiter.Allow()
		if limiter.Allow() {
			t.Error("Second request over limit should not be allowed")
		}
	})

	// ... more tests
}
