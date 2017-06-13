package extractor

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// Receiver provides an interface to allow arbitrary collection utilities
// defined within this package, or externally
type Receiver interface {
	Collect(*Candlestick) error
	Close()
}

// StdoutRcv implements Receiver to allow it to be used in a collector
type StdoutRcv struct {
	Mutex *sync.Mutex
}

// NewStdoutReceiver build a stdout Receiver
func NewStdoutReceiver(path string) *StdoutRcv {
	return &StdoutRcv{
		Mutex: &sync.Mutex{},
	}
}

// Collect prints the Candlestick to sstdout
func (r *StdoutRcv) Collect(c *Candlestick) error {
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

// CSVRcv implements Receiver to allow it to be used in a collector
type CSVRcv struct {
	Path    string
	Pointer *os.File
	Writer  *csv.Writer
	Mutex   *sync.Mutex
}

// NewCSVReceiver build a csv Receiver, cretating a blank file. existing files will be overwritten
func NewCSVReceiver(path string) (*CSVRcv, error) {
	ptr, err := os.Create(path)
	if err != nil {
		return &CSVRcv{}, err
	}

	wtr := csv.NewWriter(ptr)

	rcv := &CSVRcv{
		Path:    path,
		Pointer: ptr,
		Writer:  wtr,
		Mutex:   &sync.Mutex{},
	}

	defer rcv.Writer.Flush()
	title := []string{"Time", "Granularity", "Low", "High", "Open", "Close", "Volume"}
	err = rcv.Writer.Write(title)
	return rcv, err
}

// Close closes the file pointer
func (r *CSVRcv) Close() {
	r.Pointer.Close()
}

// Collect writes the Candlestick to the output file
func (r *CSVRcv) Collect(c *Candlestick) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	defer r.Writer.Flush()
	t := string(c.Datetime)
	g := strconv.Itoa(c.Granularity)
	l := strconv.FormatFloat(c.Low, 'E', -1, 64)
	h := strconv.FormatFloat(c.High, 'E', -1, 64)
	o := strconv.FormatFloat(c.Open, 'E', -1, 64)
	cl := strconv.FormatFloat(c.Close, 'E', -1, 64)
	v := strconv.FormatFloat(c.Volume, 'E', -1, 64)
	row := []string{t, g, l, h, o, cl, v}
	err := r.Writer.Write(row)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// JSONRcv implements Receiver to allow it to be used in a collector
type JSONRcv struct {
	Path    string
	Pointer *os.File
	Mutex   *sync.Mutex
}

// NewJSONReceiver build a json Receiver, cretating a blank file. existing files will be overwritten
func NewJSONReceiver(path string) (*JSONRcv, error) {
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
func (r *JSONRcv) Collect(c *Candlestick) error {
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
func (r *JSONRcv) Close() {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	defer r.Pointer.Close()
	r.Pointer.WriteString("]")
}

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
