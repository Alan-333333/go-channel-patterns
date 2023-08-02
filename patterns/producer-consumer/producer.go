// package producerconsumer implements a producer and consumer model
// to handle data generation and processing in goroutines.
package producerconsumer

import (
	"context"
	"sync"
	"time"
)

// Producer generates data and writes to a buffered channel.
// It controls a number of goroutines that invoke the custom ProduceFunc
// to generate data and handle errors.
type Producer struct {

	// Buffer is the buffered channel for holding produced data before
	// it is consumed.
	Buffer chan interface{}

	// NumProcs controls the number of goroutines that invoke ProduceFunc
	// to generate data concurrently.
	NumProcs int

	// ProduceFunc is the custom data generation function provided by
	// clients. It should return generated data and any error encountered.
	ProduceFunc func() (interface{}, error)

	// ErrHandler handles any errors returned by ProduceFunc.
	// If not set, errors will be ignored.
	ErrHandler func(error)

	// Notifier sends notifications about the producer lifecycle,
	// e.g. when data generation starts and finishes.
	// This can be used to add monitoring and logging.
	Notifier func(string)
}

// NewProducer creates a new Producer instance.
// It accepts bufferSize and numProcs arguments to configure the Producer.
//
// bufferSize controls the size of the buffered channel used by Producer.
// numProcs controls the number of goroutines that will generate data.
//
// A configured Producer instance is returned ready to Start data generation.
//
// Example usage:
//
//   p := NewProducer(100, 4) // bufferSize 100, four goroutines
//   p.Start()
//
func NewProducer(bufferSize int, numProcs int) *Producer {

	// Create a buffered channel to hold produced data.
	// bufferSize determines the max number of items the channel can hold
	buffer := make(chan interface{}, bufferSize)

	// Create the Producer instance with the injected config.
	p := &Producer{
		Buffer:   buffer,
		NumProcs: numProcs,
	}

	return p
}

// Run starts the Producer goroutines to generate data.
//
// It will start NumProcs goroutines to execute the ProduceFunc
// concurrently to generate data.
//
// The provided context can be used to control cancelation and timeouts.
//
// Run blocks until all goroutines finish or the context is canceled.
//
// Example usage:
//
//   ctx := context.Background()
//   p.Run(ctx)
//
func (p *Producer) Run(ctx context.Context) {

	var wg sync.WaitGroup

	// Add a WaitGroup counter for each goroutine
	for i := 0; i < p.NumProcs; i++ {
		wg.Add(1)
		// Start a goroutine to execute the ProduceFunc
		go p.runProc(ctx, &wg)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

// runProc executes the custom ProduceFunc to generate data.
// It runs in a goroutine started by the Run method.
func (p *Producer) runProc(ctx context.Context, wg *sync.WaitGroup) {

	defer wg.Done()

	for {
		// Check for context cancellation
		if p.isCancelled(ctx) {
			return
		}

		// Invoke custom function to generate data
		data, err := p.ProduceFunc()

		// Handle any errors
		if err != nil {
			p.handleError(err)
			continue
		}

		// No data produced
		if data == nil {
			return
		}
		// Write data to buffer, applying backpressure if full
		written := p.tryWrite(p.Buffer, data)
		if !written {
			p.applyBackpressure()
		}
	}
}

// Inject pipes data from the Producer's buffer channel to the
// provided out channel.

// It runs until the out channel is closed or the context is cancelled.

// Inject writes to out channel in a non-blocking manner, dropping messages
// if the out channel is full.
func (p *Producer) Inject(ctx context.Context, out chan interface{}) {

	for {

		// Check for context cancellation
		if p.isCancelled(ctx) {
			return
		}

		// Non-blocking write to out channel
		// Try read from buffer
		data, ok := p.tryReadBuffer()
		if !ok {
			return
		}

		// Try write to out channel
		written := p.tryWrite(out, data)
		if !written {
			continue
		}
		// Notify data injected
		p.Notifier("InjectFinished")

	}
}

// Close closes the Producer's buffer channel.
//
// This signals to any goroutines waiting to send to the channel that
// no more data will arrive. They will exit once the channel is drained.
//
// It should be called when data generation is complete and before disposing
// the Producer object.
func (p *Producer) Close() {

	close(p.Buffer)

}

// HandleError sets a custom error handler function.
//
// The handler will be invoked whenever an error is returned
// by the ProduceFunc during data generation.
//
// If no custom handler is set, errors will be ignored.
//
// Example:
//
//   p.HandleError(func(err error){
//     log.Printf("data generation failed: %v", err)
//   })
func (p *Producer) HandleError(handler ErrHandler) {

	p.ErrHandler = handler

}

// Notify sets a notifier function to receive lifecycle notifications.
//
// The notifier will be invoked during key events like start/stop of
// data generation.
//
// This allows monitoring systems to track the Producer lifecycle events.
//
// Example:
//
//   p.Notify(func(msg string){
//     log.Println(msg)
//   })
func (p *Producer) Notify(notifier Notifier) {

	p.Notifier = notifier

}

// Helper methods

// tryReadBuffer tries to read data from the buffer channel in a
// non-blocking way.
// It returns the data if read succeeded, otherwise nil.
// The second return value indicates if read succeeded.
func (p *Producer) tryReadBuffer() (interface{}, bool) {
	// 非阻塞读取 buffer
	select {
	case data := <-p.Buffer:
		return data, true
	default:
		return nil, false
	}
}

// tryWrite attempts to write data to out channel in non-blocking manner.
// Returns true if write succeeded, false otherwise.
func (p *Producer) tryWrite(out chan interface{}, data interface{}) bool {

	select {
	case <-out:
		// out is closed
		return false

	case out <- data:
		// Write succeeded
		return true

	default:
		// out is full
		return false
	}

}

// isCancelled checks if the context has been cancelled.
// This allows goroutines to stop when a cancellation signal is received.
func (p *Producer) isCancelled(ctx context.Context) bool {

	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}

}

// applyBackpressure applies throttling when buffer channel is full.
// This gives time for the channel to drain and prevent Producer from
// overwhelming downstream consumers.
//
// Current implementation simply sleeps for a short period. More
// sophisticated throttling and metrics can be added, for example:
//
//   - exponential backoff
//   - adaptive throttling based on consumer speed
//   - metrics for drop count, throttle time, etc
func (p *Producer) applyBackpressure() {

	// Notify backpressure applied
	p.Notifier("buff full sleep")

	// Simple throttling sleep
	time.Sleep(100 * time.Millisecond)

}

// handleError handles any errors returned by the ProduceFunc.
//
// It logs the error and invokes any configured ErrHandler.
//
// If no ErrHandler set, the error will be ignored.
//
// This method allows customizing error handling logic.
func (p *Producer) handleError(err error) {

	// Notify error happened
	p.Notifier("ProducerError")

	// Invoke custom error handler
	if p.ErrHandler != nil {
		p.ErrHandler(err)
	}

	// Any other error handling logic...

}
