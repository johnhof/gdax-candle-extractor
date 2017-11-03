package receivers

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