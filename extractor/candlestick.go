package extractor

import exchange "github.com/preichenberger/go-coinbase-exchange"

// Candlestick is a representation of trades that ocurred in a block of time.
// this redirection is necessary to simplify buffer usage with a string-type time,
// And to add granularity to the set of tracked data
type Candlestick struct {
	Datetime    string  `json:"datetime"`
	Granularity int     `json:"granularity"`
	Low         float64 `json:"low"`
	High        float64 `json:"high"`
	Open        float64 `json:"open"`
	Close       float64 `json:"close"`
	Volume      float64 `json:"volume"`
	Timestamp   int64   `json:"timestamp"`
}

// CandleFromRate takes the granularity int and historic rate and converts it to a candlestick struct
func CandleFromRate(granularity int, rt *exchange.HistoricRate) Candlestick {
	utc := rt.Time.UTC()
	return Candlestick{
		Datetime:    utc.String(),
		Granularity: granularity,
		Low:         rt.Low,
		High:        rt.High,
		Open:        rt.Open,
		Close:       rt.Close,
		Volume:      rt.Volume,
		Timestamp:   utc.Unix(),
	}
}

// CandlesFromRates takes the granularity int and list of historic rates and converts them to a list of candlestick structs
func CandlesFromRates(granularity int, rts []exchange.HistoricRate) []Candlestick {
	cdls := make([]Candlestick, len(rts))
	for i, rt := range rts {
		cdls[i] = CandleFromRate(granularity, &rt)
	}
	return cdls
}
