// ratelimit/leaky_bucket/bucket.go

package leakybucket

import (
	"testing"
	"time"
)

func TestLeakyBucket(t *testing.T) {

	t.Run("New", func(t *testing.T) {
		capacity := 100
		rate := 10
		b := New(capacity, rate)

		if b.capacity != capacity {
			t.Errorf("Capacity %d not set correctly", capacity)
		}
		if b.rate != float64(rate) {
			t.Errorf("Rate %d not set correctly", rate)
		}
	})

	t.Run("Allow", func(t *testing.T) {
		b := New(10, 1)
		for i := 0; i < 2; i++ {
			if !b.Allow() {
				t.Error("Request within capacity should pass")
			}
			time.Sleep(1 * time.Second)
		}

		// Reset bucket
		b = New(10, 1)

		if !b.Allow() {
			t.Error("First request should always pass")
		}

		for i := 0; i < 10; i++ {
			if b.Allow() {
				t.Error("Requests over capacity should be limited")
			}
		}
	})

}
