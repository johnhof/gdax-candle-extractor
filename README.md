# GDAX Candlestick Extractor

This package is designed to simplify historical [candlestick](http://www.investopedia.com/terms/c/candlestick.asp) data extraction from the GDAX API, for the purposes of analysis and  model training.

The extractor itself is abstracted to allow use in existing golang codebases in addition to the command line util.

This utility supports outputting data to the following collection tools:
  * stdout
  * CSV file
  * JSON file
  * Newline delimited JSON file
  * Elasticsearch index

`go get github.com/johnhof/gdax-candle-extractor`

## Command line usage

`$ gdax-candle-extractor --key=KEY --secret=SECRET --passphrase=PASSPHRASE --product=PRODUCT [<flags>]`

**Get candlestick data for each hour since the beginning of 01/01/17, and pipe it to a csv file**

`$ gdax-candle-extractor -start=2017-01-01T00:00:09+00:00 -granularity=3600 -out-type=csv -out-path=./data.csv`

## Docker usage

Either 

`git clone git@github.com:johnhof/gdax-candle-extractor.git && docker build gdax-candle-extractor -t extractor`

Or

``

```bash
docker run \
-e GDAX_API_KEY=foo_key \
-e GDAX_API_SECRET=foo_secret \
-e GDAX_API_PASSPHRASE=foo_phrase \
-e GDAX_EXTRACTOR_PRODUCT=ETH-USD \
-e GDAX_EXTRACTOR_VERBOSE=true \
extractor 

```

### Options

The following options are the result of `--help`. The text is modified to include environment var alternatives which will override the defaults, but not command line params.

```bash
      --help                                                                context-sensitive help (also try --help-long and --help-man).
  -v, --verbose,          GDAX_EXTRACTOR_VERBOSE                            verbose logging
  -k, --key,              GDAX_API_KEY=KEY                                  GDAX API key
  -s, --secret,           GDAX_API_SECRET=SECRET                            GDAX API secret
  -p, --passphrase,       GDAX_API_PASSPHRASE=PASSPHRASE                    GDAX API passphrase
      --product,          GDAX_EXTRACTOR_PRODUCT=PRODUCT                    Product ID to extract [BTC-USD, ETH-USD, LTC-USD]
  -G, --granularity,      GDAX_EXTRACTOR_GRANULARITY=86400                  Granularity in seconds of blocks in the candlestick data
  -b, --buffer-size,      GDAX_EXTRACTOR_BUFFER_SIZE=100                    Size of candlestick buffer waiting for collection
  -S, --start,            GDAX_EXTRACTOR_START="2017-10-31T00:11:58-07:00"  Start time as RFC3339
  -E, --end,              GDAX_EXTRACTOR_END="2017-11-06T23:11:58-08:00"    End time in as RFC3339
      --out-stdout,       GDAX_EXTRACTOR_OUT_STDOUT                         Write output to stdout. Used by default if no other output is specified
      --out-csv,          GDAX_EXTRACTOR_OUT_CSV                            Write output to CSV file
      --out-csv-file,     GDAX_EXTRACTOR_OUT_CSV_FILE="out.csv"             Set the file to write to
      --out-json,         GDAX_EXTRACTOR_OUT_JSON                           Write output to JSON file
      --out-json-file,    GDAX_EXTRACTOR_OUT_JSON_FILE="out.json"           Set the file to write to
      --out-nd-json,      GDAX_EXTRACTOR_OUT_ND_JSON                        Write output to new line delimited JSON file
      --out-nd-json-file, GDAX_EXTRACTOR_OUT_ND_JSON_FILE="out.ndjson"      Set the file to write to
      --out-es,           GDAX_EXTRACTOR_OUT_ES                             Index output to elasticsearch
      --out-es-index,     GDAX_EXTRACTOR_OUT_ES_INDEX="candlestick"         Elasticsearch index to use for output
      --out-es-host,      GDAX_EXTRACTOR_OUT_ES_HOST="localhost"            Set the elasticsearch host to write to
      --out-es-port,      GDAX_EXTRACTOR_OUT_ES_PORT="9200"                 Set the elasticsearch port to write to
      --out-es-secure,    GDAX_EXTRACTOR_SECURE                             Set the elasticsearch requests to use https
      --version                                                             Show application version.

```

## Programmatic usage

The source is comprised of two discrete steps
- Extraction
  - Create an extractor
  - Run the extractor
    - Begins extraction using the provided config
    - Pushes candlestick data onto the candlestick channel
    - Pushed retrieval errors onto the error channel

- Collection
  - Either
    - Build your own logic reading Candlestick's directly from the cannel
    - Use the provided collector
      - Either
        - Build a custom receiver and add it to the collector
        - Use one of the provided receivers

```go
package main

import (
  myReceivers "./custom-receivers"
  "github.com/johnhof/gdax-candle-extractor/extractor"
  "github.com/johnhof/gdax-candle-extractor/receivers"
}

func main() {
  // Create the extractor
	extract := extractor.New(&extractor.ExtractorConfig{
		Key:        "SuperSecretGDAXKey",
		Secret:     "SuperSecretGDAXSecret",
    Passphrase: "SuperSecretGDAXPassphrase",
		Extraction: &extractor.ExtractionConfig{
      Product:     "BTC-USD", // Bitcoin price in US dollars
      Granularity: 5, // candlesticks split by  5 second chunks
      Start:       time.Now().Sub(24*time.Hour), // from yesterday
      End:         time.Now(), // until now
		},
	})

  // Start extracting
  err := extract.Extract()
  if err != nil {
    panic(err)
  }

  // Create a collector to simplify candlestick collection from the extractor
	c := extractor.NewCollector(&extractor.CollectorConfig{
		Extractor: extract,
	})

	// Create a CSV receiver to write to
	csv, err := receivers.NewCSV("out.csv")
	if err != nil {
		panic(err)
	}
	c.Add(csv)

  // Add your custom receiver which implements extractor.Receiver
  c.Add(&myReceivers.Foo{})

  // Start pulling data from the extractor channels, and forwarding it to all receivers
  // Errors will be printed to stoud unless a handler is set
  c.Collect()
}
```

## Authors

* [John Hofrichter](github.com/johnhof)
