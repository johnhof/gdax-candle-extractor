package extractor

import (
	"fmt"
	"io"
	"sync"
)

// Collector provides and interface to collect to N collectors froma  provided pipe
type Collector struct {
	Verbose   bool
	Pipe      *CandlePipe
	Receivers []Receiver
	WaitGroup *sync.WaitGroup
}

// NewCollector builds a collector with the provided pipe, and using any receivers provided
func NewCollector(p *CandlePipe, rcvs ...Receiver) *Collector {
	return &Collector{
		Pipe:      p,
		Receivers: rcvs,
		WaitGroup: &sync.WaitGroup{},
	}
}

// Add adds the receiver to the list of reveivers to be used when the collection fires
func (c *Collector) Add(r Receiver) {
	c.Receivers = append(c.Receivers, r)
}

// Collect collects from either the collectors pipe, or the pipe param, if provided
func (c *Collector) Collect(cps ...*CandlePipe) error {
	defer c.Close()
	var p *CandlePipe
	if len(cps) == 1 {
		p = cps[0]
	} else if len(cps) > 1 {
		panic(fmt.Sprintf("Collect was given [%d] pipes. 0-1 are accepted", len(cps)))
	} else {
		p = c.Pipe
	}

	if len(c.Receivers) == 0 {
		panic("No receivers set for the collector when Collect was called")
	}

	for {
		done, err := c.CollectOne(p)
		if err != nil {
			fmt.Println(err)
		}
		if done {
			break
		}
	}

	return nil
}

// CollectOne reads one Candlestick out of the pipe and collects it to each receiver each in its own goroutine
func (c *Collector) CollectOne(p *CandlePipe) (eof bool, err error) {
	if len(c.Receivers) == 0 {
		panic("No receivers set for the collector when CollectOne was called")
	}

	cdl := &Candlestick{}
	err = p.Read(cdl)
	if err != nil {
		if err == io.EOF {
			return true, nil
		}
		return false, err
	} else if c.Verbose {
		fmt.Print(".")
	}

	for i := range c.Receivers {
		// Make sure to track active receiver collection actions. This prevents early proc termination
		c.WaitGroup.Add(1)
		go func(rcv Receiver) {
			defer c.WaitGroup.Done()
			cErr := rcv.Collect(cdl)
			if cErr != nil {
				fmt.Println(cErr) // TODO: better error handling
			}
		}(c.Receivers[i])
	}

	return false, nil
}

// Close closes all receivers
func (c *Collector) Close() {
	// Wait for receiver action to complete
	c.WaitGroup.Wait()

	// Close all receivers
	for i := range c.Receivers {
		c.Receivers[i].Close()
	}
}
