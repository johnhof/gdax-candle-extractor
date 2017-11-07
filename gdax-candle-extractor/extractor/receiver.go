package extractor

// Receiver provides an interface to allow arbitrary collection utilities
// defined within this package, or externally
type Receiver interface {
	Collect(*Candlestick) error
	Close()
}