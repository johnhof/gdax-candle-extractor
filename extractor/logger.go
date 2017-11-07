package extractor

// Logger is a an abstraction to allow injection of a logger for debugging
type Logger interface {
	Printf(format string, v ...interface{})
}
