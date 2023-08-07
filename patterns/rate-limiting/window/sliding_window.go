package window

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// SlidingWindow implements a fixed-size sliding window for rate limiting.
type SlidingWindow struct {
	sync.Mutex

	// windowSize is the size of the sliding window in time units.
	windowSize time.Duration

	// bucketSize is the size of each bucket in the window in time units.
	bucketSize time.Duration

	// bucketCount is the number of buckets in the window.
	bucketCount int

	// buckets tracks the count in each bucket.
	buckets []int

	// startTime records the start time of the window
	startTime time.Time

	// lastRequestTime records the end time of the window
	lastRequestTime time.Time
}

// NewSlidingWindow creates a new sliding window with the given window size, bucket size
// and bucket count. Window size must be divisible by bucket size.
func New(windowSize, bucketSize time.Duration, bucketCount int) (*SlidingWindow, error) {
	if windowSize%bucketSize != 0 {
		return nil, fmt.Errorf("window size must be divisible by bucket size")
	}

	if bucketCount <= 0 {
		return nil, fmt.Errorf("bucket count must be positive")
	}

	sw := &SlidingWindow{
		windowSize:  windowSize,
		bucketSize:  bucketSize,
		bucketCount: bucketCount,
		buckets:     make([]int, bucketCount),
	}
	return sw, nil
}

// Allow reports whether a new event should be allowed, and if so increments the
func (sw *SlidingWindow) Allow() bool {

	sw.Lock()
	defer sw.Unlock()

	now := time.Now()

	// Initialize start time
	if sw.startTime.IsZero() {
		sw.startTime = now
	}

	// Check if request time exceeds window size
	if now.Sub(sw.startTime) > sw.windowSize {

		// Reset window if exceeded
		sw.resetWindow(now)
		return false
	}

	// Calculate bucket index
	bucketIdx := sw.getBucketIndex(now)

	// Increment bucket count
	sw.buckets[bucketIdx]++

	// Update last request time
	sw.lastRequestTime = now

	return true

}

// Reset window by clearing buckets and resetting start time
func (sw *SlidingWindow) resetWindow(now time.Time) {
	sw.startTime = now
	sw.lastRequestTime = now

	// Clear all bucket counts
	for i := 0; i < len(sw.buckets); i++ {
		sw.buckets[i] = 0
	}
}

// Count returns the total count for the given duration
func (sw *SlidingWindow) Count(d time.Duration) int {

	sw.Lock()
	defer sw.Unlock()
	// duration over windowSize
	if d > sw.windowSize {
		return 0
	}
	// Get start bucket
	start := sw.getBucketIndex(sw.startTime)
	start = (start + sw.bucketCount) % sw.bucketCount

	// Get end bucket
	end := sw.getBucketIndex(sw.startTime.Add(d))
	if end > sw.bucketCount {
		end -= sw.bucketCount
	}

	var count int
	for i := start; i < end; i++ {
		bucketIndex := i % sw.bucketCount
		count += sw.buckets[bucketIndex]
	}

	return count
}

// Get bucket index for the given timestamp
func (sw *SlidingWindow) getBucketIndex(time time.Time) int {

	elapsed := time.Sub(sw.startTime).Nanoseconds()

	// Convert seconds to int64 to avoid precision loss
	secs := int64(elapsed)

	// Check overflow
	if secs > math.MaxInt64 {
		// Handle overflow case
		return 0
	}

	bucketSizeSecs := int64(sw.bucketSize.Nanoseconds())

	// Compute bucket index
	idx := int(secs / bucketSizeSecs)

	return idx
}

// Reset resets the counts in all buckets to 0.
func (sw *SlidingWindow) Reset() {
	sw.Lock()
	defer sw.Unlock()

	for i := 0; i < sw.bucketCount; i++ {
		sw.buckets[i] = 0
	}
}

// BucketCount returns the current count for the given bucket index.
func (sw *SlidingWindow) BucketCount(idx int) int {
	sw.Lock()
	defer sw.Unlock()

	return sw.buckets[idx%sw.bucketCount]
}
