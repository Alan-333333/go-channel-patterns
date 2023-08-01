// package Consumerconsumer implements a Consumer and consumer model
// to handle data generation and processing in goroutines.
package producerconsumer

import (
	"context"
	"sync"
	"time"
)

// ErrHandler Consumer err
type ErrHandler func(error)

// Notifier
type Notifier func(string)

// Consumer represents a consumer for processing data concurrently.
type Consumer struct {

	// Buffer is a buffered channel that holds data to be consumed.
	Buffer chan interface{}

	// NumProcs is the number of concurrent goroutines that will
	// process data.
	NumProcs int

	// ConsumeFunc is a handler function that will be invoked for
	// each data item to process it.
	ConsumeFunc func(interface{}) error

	// ErrHandler is a handler function that will be called
	// if an error occurs during processing.
	ErrHandler func(error)

	// Notifier is a callback function that will be invoked
	// on specific events.
	Notifier func(string)
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(bufferSize int, numProcs int) *Consumer {

	// Create a buffered channel to serve as the buff.
	buff := make(chan interface{}, bufferSize)

	// Initialize a Consumer instance.
	c := &Consumer{
		Buffer:   buff,
		NumProcs: numProcs,
	}

	return c
}

// Run starts the consumer by spinning up multiple
// concurrent goroutines to process data.
func (c *Consumer) Run(ctx context.Context) {

	// wg is used to wait for all goroutines to finish.
	var wg sync.WaitGroup

	// Spin up a goroutine for each processor.
	for i := 0; i < c.NumProcs; i++ {

		wg.Add(1)

		// Notify that a processor has started.
		c.Notifier("ConsumerStarted")

		// Launch a goroutine to process data.
		go c.runProc(ctx, &wg)
	}

	// Block until all processors have finished.
	wg.Wait()
}

// runProc runs in a goroutine to process data from the inbox channel.
func (c *Consumer) runProc(ctx context.Context, wg *sync.WaitGroup) {

	// Defer marking this goroutine as done in the WaitGroup.
	defer wg.Done()

	// Define a timeout duration.
	timeout := 2 * time.Second

	for {
		// ctx Timeout case
		if c.isCancelled(ctx) {
			return
		}
		// Timeout case
		if c.isTimedOut(timeout) {
			return
		}
		// Try read from buffer
		data, ok := c.tryReadBuffer()
		if !ok {
			return
		}
		// Invoke custom function to consume data
		err := c.ConsumeFunc(data)

		// Handle error
		if err != nil {
			c.handleError(err)
		}
	}

}

// Close gracefully closes the Consumer.
// It closes the Buffer channel and notifies shutdown.
func (c *Consumer) Close() {

	// Close Buffer channel
	close(c.Buffer)

}

// sets the error handler function.
func (c *Consumer) HandleError(handler ErrHandler) {
	c.ErrHandler = handler
}

// sets the Notify handler function.
func (c *Consumer) Notify(notifier Notifier) {
	c.Notifier = notifier
}

// Helper methods

// isCancelled checks if the context has been cancelled.
// This allows goroutines to stop when a cancellation signal is received.
func (c *Consumer) isCancelled(ctx context.Context) bool {

	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}

}

// handleError handles any errors returned by the ConsumeFunc.
//
// It logs the error and invokes any configured ErrHandler.
//
// If no ErrHandler set, the error will be ignored.
//
// This method allows customizing error handling logic.
func (c *Consumer) handleError(err error) {

	// Notify error happened
	c.Notifier("ConsumerError")

	// Invoke custom error handler
	if c.ErrHandler != nil {
		c.ErrHandler(err)
	}

	// Any other error handling logic...

}

// tryReadBuffer tries to read data from the buffer channel in a
// non-blocking way.
// It returns the data if read succeeded, otherwise nil.
// The second return value indicates if read succeeded.
func (c *Consumer) tryReadBuffer() (interface{}, bool) {
	// 非阻塞读取 buffer
	select {
	case data := <-c.Buffer:
		return data, true
	default:
		return nil, false
	}
}

// isTimedOut checks if timeout has occurred.
func (c *Consumer) isTimedOut(timeout time.Duration) bool {

	select {
	case <-time.After(timeout):
		return true

	default:
		return false
	}

}
