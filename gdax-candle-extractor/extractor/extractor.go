package extractor

import (
	"errors"
	"fmt"
	"time"

	exchange "github.com/preichenberger/go-coinbase-exchange"
)

// Extractor encapsulates the extracting configuration and process
type Extractor struct {
	Client          *exchange.Client
	Config          *ExtractorConfig
	Logger          Logger
	CandlestickChan chan *Candlestick
	ErrorChan       chan error
	running         bool
}

// ExtractorConfig provides values for the extractor-GDAX request configuration
type ExtractorConfig struct {
	Key        string
	Secret     string
	Passphrase string
	Logger     Logger
	BufferSize int
	Extraction *ExtractionConfig
}

// ExtractionConfig providesvalues for the actual extracting execution
type ExtractionConfig struct {
	Product     string
	Start       time.Time
	End         time.Time
	Granularity int
}

// New builds an initialized extractor
func New(config *ExtractorConfig) *Extractor {
	client := exchange.NewClient(config.Secret, config.Key, config.Passphrase)
	return &Extractor{
		Client: client,
		Config: config,
		Logger: config.Logger,
	}
}

// Start gets trade history and writes each result to the channels. extraction buckets are split by `nil`
func (m *Extractor) Start() error {
	if m.running == true {
		return errors.New("Extractor already started")
	}
	rngs := buildReqRanges(m.Config.Extraction)

	// sleep time is used to wait between requests and evade ratelimiting
	waitMin := 400 * time.Millisecond

	m.CandlestickChan = make(chan *Candlestick, m.Config.BufferSize)
	m.ErrorChan = make(chan error, m.Config.BufferSize)

	// Make a request every .4 seconds(to avoid ratelimiting). pipe the output to the collectors
	m.running = true
	go func() {
		for _, rng := range rngs {
			if !m.running {
				break
			}
			start := rng[0]
			end := rng[1]
			started := time.Now()

			// Make req
			cdls, err := m.GetCandleRange(m.Config.Extraction.Product, start, end, m.Config.Extraction.Granularity)
			if err != nil {
				m.ErrorChan <- err
			}

			// Log if set
			if m.Config.Logger != nil {
				tRng := fmt.Sprintf("(%s - %s)", start.String(), end.String())
				tDif := end.Sub(start).String()
				m.Logger.Printf("\n=> REQ: [%s:%d] %s=%s\n<= RES: %d results\n", m.Config.Extraction.Product, m.Config.Extraction.Granularity, tRng, tDif, len(cdls))
			}

			// Send results
			for _, cdl := range cdls {
				func(candle Candlestick) {
					m.CandlestickChan <- &candle
				}(cdl)
			}

			// sleep until we reached an acceptable rate according to the GDAX API
			time.Sleep(waitMin - (time.Since(started) / time.Millisecond))
		}
		if m.running {
			m.Stop()
		}
	}()
	return nil
}

// Stop closed the channels and snds the extraction process
func (m *Extractor) Stop() {
	m.running = false
	close(m.CandlestickChan)
	close(m.ErrorChan)
}

// Candlesticks returns the candlestick channel
func (m *Extractor) Candlesticks() chan *Candlestick {
	return m.CandlestickChan
}

// Errors returns the error channel
func (m *Extractor) Errors() chan error {
	return m.ErrorChan
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
		return cdls, fmt.Errorf("GDAX Request Error: [%s] %s", product, err.Error())
	}

	return CandlesFromRates(granularity, rts), nil
}

//
// Internal helpers
//

// buildReqRanges takes the extracting config and breaks it into 200 result-request blocks
// To maintain compliance with the bounds of the GDAX API
func buildReqRanges(config *ExtractionConfig) [][]time.Time {
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
