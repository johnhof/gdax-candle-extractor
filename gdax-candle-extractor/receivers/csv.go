package receivers

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

// Close closes the file pointer
func (r *CSVRcv) Close() {
	r.Pointer.Close()
}
