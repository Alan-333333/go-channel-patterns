package window

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Case 1: window size is divisible by bucket size.
	windowSize := 10 * time.Second
	bucketSize := 2 * time.Second
	bucketCount := 5
	sw, err := New(windowSize, bucketSize, bucketCount)
	if err != nil {
		t.Errorf("New() failed: %v", err)
	}
	if sw.windowSize != windowSize {
		t.Errorf("sw.windowSize = %v, want %v", sw.windowSize, windowSize)
	}
	if sw.bucketSize != bucketSize {
		t.Errorf("sw.bucketSize = %v, want %v", sw.bucketSize, bucketSize)
	}
	if sw.bucketCount != bucketCount {
		t.Errorf("sw.bucketCount = %v, want %v", sw.bucketCount, bucketCount)
	}

	// Case 2: window size is not divisible by bucket size.
	windowSize = 11 * time.Second
	bucketSize = 2 * time.Second
	bucketCount = 5
	_, err = New(windowSize, bucketSize, bucketCount)
	if err == nil {
		t.Errorf("New() should have failed")
	}

	// Case 3: bucket count is not positive.
	windowSize = 10 * time.Second
	bucketSize = 2 * time.Second
	bucketCount = -1
	_, err = New(windowSize, bucketSize, bucketCount)
	if err == nil {
		t.Errorf("New() should have failed")
	}
}

func TestAllow(t *testing.T) {
	// Case 1: allow new event.
	windowSize := 100 * time.Millisecond
	bucketSize := 2 * time.Millisecond
	bucketCount := 50
	sw, err := New(windowSize, bucketSize, bucketCount)
	if err != nil {
		t.Errorf("New() failed: %v", err)
	}

	ok := sw.Allow()
	if !ok {
		t.Errorf("Allow() should have returned true")
	}

	// Case 2: reject new event because window has been exceeded.
	time.Sleep(windowSize + bucketSize)
	ok = sw.Allow()
	if ok {
		t.Errorf("Allow() should have returned false")
	}
}

func TestResetWindow(t *testing.T) {
	// Case 1: reset window.
	windowSize := 10 * time.Second
	bucketSize := 2 * time.Second
	bucketCount := 5
	sw, err := New(windowSize, bucketSize, bucketCount)
	if err != nil {
		t.Errorf("New() failed: %v", err)
	}

	now := time.Now()
	sw.Allow()
	sw.resetWindow(now)

	for i := 0; i < len(sw.buckets); i++ {
		if sw.buckets[i] != 0 {
			t.Errorf("sw.buckets[%v] = %v, want 0", i, sw.buckets[i])
		}
	}
}

func TestCount(t *testing.T) {

	sw, _ := New(10*time.Second, 1*time.Second, 10) // 创建测试滑动窗口

	// 1个时间单位内计数
	sw.buckets[0] = 1
	c := sw.Count(sw.bucketSize)
	if c != 1 {
		t.Errorf("count 1 time unit failed, got %d", c)
	}

	// 多个时间单位内计数
	sw.buckets[0] = 1
	sw.buckets[1] = 2
	c = sw.Count(2 * sw.bucketSize)
	if c != 3 {
		t.Errorf("count 2 time units failed, got %d", c)
	}

	sw.buckets[2] = 5
	c = sw.Count(sw.windowSize)
	if c != 8 {
		t.Errorf("count 2 time units failed, got %d", c)
	}

	// 超过时间范围的计数
	c = sw.Count(12 * sw.bucketSize)
	if c != 0 {
		t.Errorf("count exceeded range should be 0, got %d", c)
	}

}

func TestGetBucketIndex(t *testing.T) {
	sw, _ := New(10*time.Second, 1*time.Second, 10) // 创建测试滑动窗口

	now := time.Now()
	idx := sw.getBucketIndex(now)

	expected := int(now.Sub(sw.startTime) / sw.bucketSize)
	if idx != expected {
		t.Errorf("Got %d, expect %d", idx, expected)
	}
}

func TestReset(t *testing.T) {
	sw := newTestSlidingWindow()

	sw.buckets[0] = 10
	sw.Reset()

	for i := 0; i < sw.bucketCount; i++ {
		if sw.buckets[i] != 0 {
			t.Error("Reset failed")
		}
	}
}
func newTestSlidingWindow() *SlidingWindow {
	// 创建滑动窗口
	sw, _ := New(10*time.Second, 1*time.Second, 10) // 创建测试滑动窗口
	return sw
}
