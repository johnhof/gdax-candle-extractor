# GDAX Candlestick Extractor

This package is designed to simplify historical candlestick data extraction from the GDAX API, for the purposes of analysis and  model training.

The extractor itself is abstracted to allow use in existing golang codebases in addition to the command line util.

This utility supports outputting data to the following collection tools:
  * stdout
  * CSV file
  * JSON file
  * Elasticsearch index

`go get github.com/johnhof/gdax-candle-extractor`

## Command line usage

`$ gdax-candle-extractor --key=KEY --secret=SECRET --passphrase=PASSPHRASE --product=PRODUCT [<flags>]`

**Get candlestick data for each hour since the beginning of 01/01/17, and pipe it to a csv file**

`$ gdax-candle-extractor -start=2017-01-01T00:00:09+00:00 -granularity=1h -out-type=csv -out-path=./data.csv`

### Options

```bash
    --help                              Show context-sensitive help (also try --help-long and --help-man).
    --version                           Show application version.

-v, --verbose                           verbose logging

-k, --key=KEY                           GDAX API key
-s, --secret=SECRET                     GDAX API secret
-p, --passphrase=PASSPHRASE             GDAX API passphrase
    --product=PRODUCT                   Product ID to extract [BTC-USD, ETH-USD, LTC-USD]
-G, --granularity=86400                 Granularity in seconds of blocks in the candlestick data
-S, --start="2017-06-05T11:46:30-07:00" start time as RFC3339
-E, --end="2017-06-12T11:46:30-07:00"   End time in as RFC3339

    --out-stdout                        Write output to stdout. Used by default if no other output is specified

    --out-csv                           Write output to CSV file
    --out-csv-file="out.csv"            Set the file to write to

    --out-json                          Write output to JSON file
    --out-json-file="out.json"          Set the file to write to

    --out-es                            Index output to elasticsearch
    --out-es-index="candlestick"        Elasticsearch index to use for output
    --out-es-host="localhost"           set the elasticsearch host to write to
    --out-es-port="9200"                set the elasticsearch port to write to
    --out-es-secure                     set the elasticsearch requests to use https
```

## Programmatic usage

The source is comprised of two discrete steps
- Extraction
  - Create an extractor
  - Run the extractor
    - Begins extraction using the provided config
    - Pushes candlestick data onto the pipe
    - CandlePipe wraps [io.Pipe()](https://golang.org/pkg/io/#Pipe)
      - No internal buffering exists
      - Writes will block until data is read

- Collection
  - Either
    - Build your own logic reading Candlestick's directly from the pipe
    - Use the provided collector
      - Either
        - Build a custom receiver and add it to the collector
        - Use N provided receivers

```go
package main

import (
  receivers "./custom-receivers"
  "github.com/johnhof/gdax-candle-extractor/extractor"
}

func main() {
  // Create the extractor
	m := extractor.New(extractor.ExtractorConfig{
		Key:        "SuperSecretGDAXKey",
		Secret:     "SuperSecretGDAXSecret",
		Passphrase: "SuperSecretGDAXPassphrase",
	})

  // Start extracting
	pipe := m.Extract(extractor.ExtractingConfig{
		Product:     "BTC-USD", // Bitcoin price in US dollars
    Granularity: 5, // Every 5 seconds
		Start:       time.Now().Sub(24*time.Hour), // from yesterday
		End:         time.Now(), // until now
	})

  // Create a collector to simplify candlestick collection from the pipe
	c := extractor.NewCollector(pipe)

	// Create a CSV receiver to write to
	csv, err := extractor.NewCSVReceiver("out.csv")
	if err != nil {
		panic(err)
	}
	c.Add(csv)

  // Add your custom receiver which implements extractor.Receiver
  c.add(&receivers.Foo{})

  // Start pulling data from the extractor pipe, and forwarding it to all receivers
  c.Collect()
}
```

## Authors

* [John Hofrichter](github.com/johnhof)
