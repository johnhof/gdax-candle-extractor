package receivers

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/johnhof/gdax-candle-extractor/extractor"
)

// JSONRcv implements Receiver to allow it to be used in a collector
type NLDJSONRcv struct {
	Path    string
	Pointer *os.File
	Mutex   *sync.Mutex
}

// NewNDLJSON build a json Receiver, cretating a blank file. existing files will be overwritten
func NewNLDJSON(path string) (*JSONRcv, error) {
	ptr, err := os.Create(path)
	if err != nil {
		return &JSONRcv{}, err
	}
	rcv := &JSONRcv{
		Path:    path,
		Mutex:   &sync.Mutex{},
		Pointer: ptr,
	}

	_, err = rcv.Pointer.WriteString("[\n")

	return rcv, err
}

// Collect writes the Candlestick to the output file
func (r *NLDJSONRcv) Collect(c *extractor.Candlestick) error {
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

	_, err = r.Pointer.WriteString(",\n")
	return err
}

// Close closes the file pointer
func (r *NLDJSONRcv) Close() {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	defer r.Pointer.Close()
	r.Pointer.WriteString("]")
}
