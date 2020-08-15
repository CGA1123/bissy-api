package pool

import (
	"log"
	"sync"
)

// Pool implements a generic worker queue to which arbitrary functions can be
// pushed to for asynchronous processing
type Pool interface {
	Add(func()) bool
	Start() Pool
	Stop()
}

// Config holds the configuration for an async implementation of a Pool
// with configurable concurrency and channel buffer.
type Config struct {
	channel     chan func()
	concurrency int
	wg          sync.WaitGroup
}

// New returns a new pool.Config ready to be started and added to
// concurrency sets the number of workers, buffer sets the maximum backlog
// before pushing to the queue would block
func New(concurrency, buffer int) *Config {
	return &Config{concurrency: concurrency, channel: make(chan func(), buffer)}
}

// Start kicks off the worker pool
func (c *Config) Start() Pool {
	for i := 0; i < c.concurrency; i++ {
		c.wg.Add(1)

		go func() {
			defer c.wg.Done()

			for work := range c.channel {
				work()
			}
		}()
	}

	return c
}

// Add pushes a function to the pool for processing
func (c *Config) Add(f func()) bool {
	select {
	case c.channel <- f:
		return true
	default:
		return false
	}
}

// Stop closes the pool and drains it
func (c *Config) Stop() {
	close(c.channel)

	log.Printf("pool stopped, queue: %v\n", len(c.channel))

	c.wg.Wait()
}
