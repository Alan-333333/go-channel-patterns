package leakybucket

import (
	"fmt"
	"time"
)

var bucket = New(100, 1000) // Global leakage bucket 1000 (REQs/sec)

func handleRequest(i int) {
	if !bucket.Allow() {
		fmt.Println("Request", i, "limited")
		return
	}

	fmt.Println("Handling request", i)
	// Processing requests...
}

func main() {
	for i := 1; i <= 20; i++ {
		handleRequest(i)
		time.Sleep(50 * time.Millisecond) // increase the interval
	}

	// Waiting for the leaky bucket to refill
	time.Sleep(1 * time.Second)

	for i := 21; i <= 30; i++ {
		handleRequest(i)
		time.Sleep(10 * time.Millisecond) // increase the interval
	}
}
