package extractor

import (
	"errors"
	"fmt"
	"sync"
)

// Collector acts as an N fanout pipe from an extractor to receivers. It also simplifies collection by abstracting channel complexity
type Collector struct {
	// Extractor is expected to pass candlesticks and errors over their respective channels
	Extractor Collectable
	// Receivers is the list of receivers that receive candlesticks over the `.Collect()` function
	Receivers []Receiver
	// Error handler syncronously passes errors from the extrator to the function. Default function prints to stdout
	ErrorHandler func(error)
	// running tracks whether or not the collecter is active
	running bool
}

// CollectorConfig encapsulates the collection configuration and process
type CollectorConfig struct {
	// Extractor is expected to pass candlesticks and errors over their respective channels
	Extractor Collectable
	// Receivers is the list of receivers that receive candlesticks over the `.Collect()` function
	Receivers []Receiver
	// Override the default error handler, which prints to stdout
	ErrorHandler func(error)
}

// Collectable provides an abstraction to allow any etractor impementation to be used
type Collectable interface {
	Candlesticks() chan *Candlestick
	Errors() chan error
	Stop()
}

// NewCollector builds a collector with the provided chan, and using any receivers provided
func NewCollector(config *CollectorConfig) *Collector {
	c := &Collector{
		Extractor: config.Extractor,
		Receivers: config.Receivers,
	}
	if config.ErrorHandler != nil {
		c.ErrorHandler = config.ErrorHandler
	} else {
		c.ErrorHandler = func(e error) {
			fmt.Printf("Extraction Error: %s\n", e.Error())
		}
	}
	return c
}

// Add adds the receiver to the list of reveivers to be used when the collection fires
func (c *Collector) Add(r Receiver) {
	c.Receivers = append(c.Receivers, r)
}

// Collect collects from either the collectors chan, or the chan param, if provided
func (c *Collector) Collect() error {
	if c.running {
		return errors.New("Collection already started")
	}
	defer c.Close()

	if len(c.Receivers) == 0 {
		return errors.New("No receivers set for the collector when Collect was called")
	}

	c.running = true
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for cdl := range c.Extractor.Candlesticks() {
			c.fanOut(cdl)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range c.Extractor.Errors() {
			c.ErrorHandler(err)
		}
	}()
	wg.Wait()
	c.Close()
	return nil
}

func (c *Collector) fanOut(cdl *Candlestick) (err error) {
	for _, rcv := range c.Receivers {
		cErr := rcv.Collect(cdl)
		if cErr != nil {
			c.ErrorHandler(cErr)
		}
	}
	return nil
}

// Close stops the extractor and closes all receivers
func (c *Collector) Close() {
	c.running = false

	// Close all receivers
	for i := range c.Receivers {
		c.Receivers[i].Close()
	}
}
