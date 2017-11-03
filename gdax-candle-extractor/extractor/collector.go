package extractor

import (
	"fmt"
)

// Collector provides and interface to collect to N collectors froma  provided pipe
type Collector struct {
	Verbose      bool
	Candlesticks *chan Candlestick
	Errors       *chan error
	Receivers    []Receiver
}

// NewCollector builds a collector with the provided chan, and using any receivers provided
func NewCollector(p *chan Candlestick, rcvs ...Receiver) *Collector {
	return &Collector{
		Candlesticks: p,
		Receivers:    rcvs,
	}
}

// Add adds the receiver to the list of reveivers to be used when the collection fires
func (c *Collector) Add(r Receiver) {
	c.Receivers = append(c.Receivers, r)
}

// Collect collects from either the collectors chan, or the chan param, if provided
func (c *Collector) Collect(cps ...*chan Candlestick) error {
	defer c.Close()
	var candlesticks *chan Candlestick
	if len(cps) == 1 {
		candlesticks = cps[0]
	} else if len(cps) > 1 {
		panic(fmt.Sprintf("Collect was given [%d] pipes. 0-1 are accepted", len(cps)))
	} else {
		candlesticks = c.Candlesticks
	}

	if len(c.Receivers) == 0 {
		panic("No receivers set for the collector when Collect was called")
	}

	for {
		candlestick <- candlesticks
		if c.Verbose {
			fmt.Print(".")
		}

	}

	return nil
}

func (c *Collector) fanOut(cdl *Candlestick) (eof bool, err error) {
	for i := range c.Receivers {
		cErr := rcv.Collect(cdl)
		if cErr != nil {
			c.Errors <- cErr
		}
	}

	return false, nil
}

// Close closes all receivers
func (c *Collector) Close() {
	close(c.Candlesticks)
	close(c.Errors)

	// Close all receivers
	for i := range c.Receivers {
		c.Receivers[i].Close()
	}
}
