package tokenbucket

import (
	"fmt"
	"time"
)

func main() {
	// Create a token bucket with rate 10 tokens per second and capacity 10 tokens.
	tb := New(10, 10)

	// Start goroutine to take tokens from the bucket.
	go func() {
		for {
			// Wait until a token becomes available.
			tb.Wait()

			// Do something with the token.
			fmt.Println("Got a token")

			// Sleep for a second.
			time.Sleep(1 * time.Second)
		}
	}()

	// Run for 10 seconds.
	time.Sleep(10 * time.Second)

	// Close the token bucket.
	tb.Close()
}
