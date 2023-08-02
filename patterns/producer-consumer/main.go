package producerconsumer

import (
	"context"
	"math/rand"
	"sync"
)

const (
	// Producer buffer size
	producerBufferSize = 1000

	// Number of producer workers
	producerWorkers = 10

	// Consumer buffer size
	consumerBufferSize = 1000

	// Number of consumer workers
	consumerWorkers = 20
)

// newProducer creates a new producer
func newProducer() *Producer {
	return NewProducer(
		producerBufferSize,
		producerWorkers,
	)
}

// newConsumer creates a new consumer
func newConsumer() *Consumer {
	return NewConsumer(
		consumerBufferSize,
		consumerWorkers,
	)
}

func main() {
	p := newProducer()
	p.ProduceFunc = func() (interface{}, error) {
		// Create rand nums
		data := rand.Intn(100)
		// Stop producing
		if data > 90 {
			return nil, nil
		}
		return data, nil
	}
	// Set Notifier and ErrHandler
	p.Notifier = func(s string) {}
	p.ErrHandler = func(err error) {}

	// Create consumer
	c := newConsumer()
	// Set consume function
	c.ConsumeFunc = func(data interface{}) error {
		// Consume consumes data
		return nil
	}
	// Set Notifier and ErrHandler
	c.Notifier = func(s string) {}
	c.ErrHandler = func(err error) {}

	// Start producer and consumer in goroutines
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	//run producer and consumer goroutines
	go func() {
		p.Run(ctx)
		wg.Done()
	}()
	go func() {
		c.Run(ctx)
		wg.Done()
	}()

	// Run data inject from producer to consumer
	p.Inject(ctx, c.Buffer)
	// Wait end
	wg.Wait()
}
