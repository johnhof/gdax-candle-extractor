package receivers

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/johnhof/gdax-candle-extractor/extractor"
)

// NDJSONRcv implements Receiver to allow it to be used in a collector
type NDJSONRcv struct {
	Path    string
	Pointer *os.File
	Mutex   *sync.Mutex
}

// NewNDJSON build a json Receiver, cretating a blank file. existing files will be overwritten
func NewNDJSON(path string) (*JSONRcv, error) {
	ptr, err := os.Create(path)
	if err != nil {
		return &JSONRcv{}, err
	}
	rcv := &JSONRcv{
		Path:    path,
		Mutex:   &sync.Mutex{},
		Pointer: ptr,
	}

	return rcv, nil
}

// Collect writes the Candlestick to the output file
func (r *NDJSONRcv) Collect(c *extractor.Candlestick) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	b, err := json.Marshal(c)
	if err != nil {
		return err
	}

	_, err = r.Pointer.Write(b)
	if err != nil {
		return err
	}

	_, err = r.Pointer.WriteString("\n")
	return err
}

// Close closes the file pointer
func (r *NDJSONRcv) Close() {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	defer r.Pointer.Close()
	r.Pointer.WriteString("]")
}
