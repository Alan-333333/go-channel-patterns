package window

import (
	"fmt"
	"time"
)

func main() {

	// Create a sliding window with a window
	// Size of 300 milliseconds, a bucket size of 100 milliseconds, and 3 buckets.
	sw, err := New(300*time.Millisecond, 100*time.Millisecond, 3)
	if err != nil {
		panic(err)
	}

	// Send 10 consecutive requests
	for i := 0; i < 10; i++ {
		sw.Allow()
		time.Sleep(20 * time.Millisecond)
		// Count the number of requests in the first 300 millseconds
		sw.Count(300 * time.Millisecond)
	}

	// Resend 10 consecutive requests
	for i := 0; i < 10; i++ {
		sw.Allow()
		time.Sleep(20 * time.Millisecond)
		sw.Count(300 * time.Millisecond)
	}

	// Reset
	sw.Reset()

	// Get no.3 bucket
	fmt.Println(sw.BucketCount(2)) //
}
