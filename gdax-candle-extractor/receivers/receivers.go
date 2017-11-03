package receivers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

// ESRcv implements Receiver to allow it ot be used in a collector
type ESRcv struct {
	Secure  bool
	Index   string
	Host    string
	Port    string
	BaseURL string
	Mutex   *sync.Mutex
	Client  *http.Client
}

// ESDocBody wraps the candlestick in an accepted json format for the upsert operation
type ESDocBody struct {
	Doc      *Candlestick `json:"doc"`
	AsUpsert bool         `json:"doc_as_upsert"`
}

// NewESReceiver build a json Receiver, cretating a blank file. existing files will be overwritten
func NewESReceiver(index string, host string, port string, secOpt ...bool) (*ESRcv, error) {
	// prt, err := os.OpenFile(filename, os.O_APPEND, 0666)
	protocol := "http"
	secure := false
	if len(secOpt) > 0 && secOpt[0] == true {
		secure = true
		protocol = "https"
	}

	rcv := &ESRcv{
		Secure:  secure,
		Index:   index,
		Host:    host,
		Port:    port,
		BaseURL: fmt.Sprintf("%s://%s:%s/%s", protocol, host, port, index),
		Mutex:   &sync.Mutex{},
		Client:  &http.Client{},
	}
	return rcv, nil
}

// Collect upserts the candlestick into the set index. the candlestick
// granularity in seconds is the type, and the time string is the ID, used to
// prevent double-indexing existing executions
func (r *ESRcv) Collect(c *Candlestick) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	b, err := json.Marshal(ESDocBody{c, true})
	if err != nil {
		return err
	}

	// add the type and ID to the upsert request
	URL := fmt.Sprintf("%s/%d/%s/_update", r.BaseURL, c.Granularity, c.Datetime)
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode > 299 {
		bts, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("ERR: [%d] %s", res.StatusCode, string(bts))
	}
	return err
}

// Close acts as a no-op to implement the receiver interface
func (r *ESRcv) Close() {
	// NOOP
}
