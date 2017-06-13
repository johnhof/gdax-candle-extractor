package extractor

import (
	"encoding/gob"
	"fmt"
	"io"
)

// CandlePipe handles piping and coding of data for communication between the extractor and colletion propcess
type CandlePipe struct {
	Reader  *io.PipeReader
	Decoder *gob.Decoder
	Writer  *io.PipeWriter
	Encoder *gob.Encoder
}

// NewCandlePipe creates an initialized CandlePipe
func NewCandlePipe() *CandlePipe {
	p := &CandlePipe{}
	p.Reader, p.Writer = io.Pipe()
	p.Decoder = gob.NewDecoder(p.Reader)
	p.Encoder = gob.NewEncoder(p.Writer)
	return p
}

// Read reads the next trade off of the pipe
func (p *CandlePipe) Read(c *Candlestick) error {
	// var i interface{}
	// err := p.Decoder.Decode(i)
	// if err != nil {
	// 	return err
	// }
	//
	// return nil
	return p.Decoder.Decode(&c)
}

// WriteAll writes the set of candlesticks into the pipe
func (p *CandlePipe) WriteAll(cs []Candlestick) (err []error) {
	var errs []error
	for _, c := range cs {
		err := p.Write(&c)
		if err != nil {
			// TODO: something better
			fmt.Println(err)
			errs = append(errs, err)
		}
	}
	return errs
}

// Write writes the candlestick into the pipe
func (p *CandlePipe) Write(c *Candlestick) (err error) {
	return p.Encoder.Encode(*c)
}

// Close closes the pipes
func (p *CandlePipe) Close() {
	p.Reader.Close()
	p.Writer.Close()
}
