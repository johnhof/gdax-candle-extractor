package receivers

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/johnhof/gdax-candle-extractor/extractor"
)

// CSVRcv implements Receiver to allow it to be used in a collector
type CSVRcv struct {
	Path    string
	Pointer *os.File
	Writer  *csv.Writer
	Mutex   *sync.Mutex
}

// NewCSV build a csv Receiver, cretating a blank file. existing files will be overwritten
func NewCSV(path string) (*CSVRcv, error) {
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

// Collect writes the Candlestick to the output file
func (r *CSVRcv) Collect(c *extractor.Candlestick) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	defer r.Writer.Flush()
	t := string(c.Datetime)
	g := strconv.Itoa(c.Granularity)
	l := fToS(c.Low)
	h := fToS(c.High)
	o := fToS(c.Open)
	cl := fToS(c.Close)
	v := fToS(c.Volume)
	row := []string{t, g, l, h, o, cl, v}
	err := r.Writer.Write(row)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// Close closes the file pointer
func (r *CSVRcv) Close() {
	r.Pointer.Close()
}

func fToS(f float64) string {
	return strings.Trim(fmt.Sprintf("%f", f), "0")
}
