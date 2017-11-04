package receivers

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/johnhof/gdax-candle-extractor/extractor"
)

// StdoutRcv implements Receiver to allow it to be used in a collector
type StdoutRcv struct {
	Mutex *sync.Mutex
}

// NewStdout build a stdout Receiver
func NewStdout(path string) *StdoutRcv {
	return &StdoutRcv{
		Mutex: &sync.Mutex{},
	}
}

// Collect prints the Candlestick to sstdout
func (r *StdoutRcv) Collect(c *extractor.Candlestick) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// Close acts as a no-op to implement the receiver interface
func (r *StdoutRcv) Close() {
	// NOOP
}
