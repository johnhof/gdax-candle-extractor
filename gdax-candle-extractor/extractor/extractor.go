package extractor

import (
	"fmt"
	"time"

	exchange "github.com/preichenberger/go-coinbase-exchange"
)

// Extractor encapsulates the extracting configuration and process
type Extractor struct {
	Client         *exchange.Client
	Config         ExtractorConfig
	CandlesticChan chan *Candlestick
	ErrorChan      chan error
}

// ExtractorConfig provides values for the extractor-GDAX request configuration
type ExtractorConfig struct {
	Key        string
	Secret     string
	Passphrase string
}

// ExtractingConfig providesvalues for the actual extracting execution
type ExtractingConfig struct {
	BufferSize  int
	Product     string
	Start       time.Time
	End         time.Time
	Granularity int
	Verbose     bool
}

// New builds an initialized extractor
func New(config ExtractorConfig) *Extractor {
	client := exchange.NewClient(config.Secret, config.Key, config.Passphrase)
	return &Extractor{
		Client: client,
		Config: config,
	}
}

// Extract gets trade history and writes each result to the returned channel
func (m *Extractor) Extract(config ExtractingConfig) (chan *Candlestick, chan error) {
	m.CandlesticChan = make(chan *Candlestick, config.BufferSize)
	m.ErrorChan = make(chan error)
	go m.ExtractToChan(m.CandlesticChan, m.ErrorChan, config)
	return m.CandlesticChan, m.ErrorChan
}

// ExtractToChan gets trade history and writes each result to the provided pipe
func (m *Extractor) ExtractToChan(c chan *Candlestick, e chan error, config ExtractingConfig) {
	rngs := buildReqRanges(config)

	// sleep time is used to wait between requests and evade ratelimiting
	waitMin := 400 * time.Millisecond

	// Make a request every .4 seconds(to avoid ratelimiting). pipe the output to the collectors
	for _, rng := range rngs {
		start := rng[0]
		end := rng[1]
		started := time.Now()
		cdls, err := m.GetCandleRange(config.Product, start, end, config.Granularity)
		if config.Verbose {
			tRng := fmt.Sprintf("(%s - %s)", start.String(), end.String())
			tDif := end.Sub(start).String()
			fmt.Printf("\n=> REQ: [%s:%d] %s=%s\n<= RES: %d results\n", config.Product, config.Granularity, tRng, tDif, len(cdls))
		}
		if err != nil {
			e <- err
		}
		for _, cdl := range cdls {
			func(candle Candlestick) {
				c <- &candle
			}(cdl)
		}

		// sleep until we reached an acceptable rate according to the GDAX API
		time.Sleep(waitMin - (time.Since(started) / time.Millisecond))
	}
}

// GetCandleRange returns a set of cnadlestick structs from the exchange for the product, range, and granularity
func (m *Extractor) GetCandleRange(product string, start time.Time, end time.Time, granularity int) ([]Candlestick, error) {
	var cdls []Candlestick
	params := exchange.GetHistoricRatesParams{
		Start:       start,
		End:         end,
		Granularity: granularity,
	}

	rts, err := m.Client.GetHistoricRates(product, params)
	if err != nil {
		fmt.Println(err)
		return cdls, err
	}

	return CandlesFromRates(granularity, rts), nil
}

// Candlesticks returns the candlestick channel
func (m *Extractor) Candlesticks() chan *Candlestick {
	return m.CandlesticChan
}

// Errors returns the error channel
func (m *Extractor) Errors() chan error {
	return m.ErrorChan
}

//
// Internal helpers
//

// buildReqRanges takes the extracting config and breaks it into 200 result-request blocks
// To maintain compliance with the bounds of the GDAX API
func buildReqRanges(config ExtractingConfig) [][]time.Time {
	var bs [][]time.Time

	// deterextract the time frame for each request
	frame := time.Duration(config.Granularity*200) * time.Second

	s := config.Start
	e := s.Add(frame)

	// While the current end date of the range is before than the target end date
	for config.End.Sub(e) > 0 {
		bs = append(bs, []time.Time{s, e})
		// Advance the frame
		s = e
		e = e.Add(frame)
	}

	// Add the final subsection of the range
	bs = append(bs, []time.Time{s, config.End})

	return bs
}
