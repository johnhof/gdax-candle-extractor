package extractor

import (
	"errors"
	"sync"
)

// Collector acts as an N fanout pipe from an extractor to receivers
type Collector struct {
	Extractor Collectable
	Receivers []Receiver
	ErrorChan chan error
	running   bool
}

// CollectorConfig encapsulates the collection configuration and process
type CollectorConfig struct {
	Extractor Collectable
	Receivers []Receiver
}

// Collectable provides an abstraction to allow any etractor impementation to be used
type Collectable interface {
	Candlesticks() chan *Candlestick
	Errors() chan error
	Stop()
}

// NewCollector builds a collector with the provided chan, and using any receivers provided
func NewCollector(config *CollectorConfig) *Collector {
	return &Collector{
		Extractor: config.Extractor,
		Receivers: config.Receivers,
	}
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
	c.ErrorChan = make(chan error)

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
			c.ErrorChan <- err
		}
	}()
	wg.Wait()
	return nil
}

func (c *Collector) fanOut(cdl *Candlestick) (err error) {
	for _, rcv := range c.Receivers {
		cErr := rcv.Collect(cdl)
		if cErr != nil {
			c.ErrorChan <- cErr
		}
	}
	return nil
}

func (c *Collector) Errors() chan error {
	return c.ErrorChan
}

// Close stops the extractor and closes all receivers
func (c *Collector) Close() {
	c.running = false
	close(c.ErrorChan)

	// Close all receivers
	for i := range c.Receivers {
		c.Receivers[i].Close()
	}
}
