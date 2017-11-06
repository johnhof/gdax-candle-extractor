package main

import (
	"fmt"
	"time"

	"github.com/johnhof/gdax-candle-extractor/extractor"
	"github.com/johnhof/gdax-candle-extractor/receivers"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type Log struct{}

func (l Log) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

var (
	now     = time.Now()
	timeFmt = time.RFC3339 // "2017-01-01T00:00:00+00:00"
	verbose = kingpin.Flag("verbose", "verbose logging").Short('v').
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_VERBOSE").
		Default("false").Bool()
	key = kingpin.Flag("key", "GDAX API key").Short('k').
		OverrideDefaultFromEnvar("GDAX_API_KEY").
		Required().String()
	secret = kingpin.Flag("secret", "GDAX API secret").Short('s').
		OverrideDefaultFromEnvar("GDAX_API_SECRET").
		Required().String()
	passphrase = kingpin.Flag("passphrase", "GDAX API passphrase").Short('p').
			OverrideDefaultFromEnvar("GDAX_API_PASSPHRASE").
			Required().String()
	product = kingpin.Flag("product", "Product ID to extract [BTC-USD, ETH-USD, LTC-USD]").
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_PRODUCT").
		Required().String()
	granularity = kingpin.Flag("granularity", "Granularity in seconds of blocks in the candlestick data").Short('G').
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_GRANULARITY").
			Default("86400").Int()
	bufferSize = kingpin.Flag("buffer-size", "Size of candlestick buffer waiting for collection").Short('b').
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_BUFFER_SIZE").
			Default("100").Int()
	start = kingpin.Flag("start", "start time as RFC3339").Short('S').
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_START").
		Default(now.Add(-24 * 7 * time.Hour).Format(timeFmt)).String()
	end = kingpin.Flag("end", "End time in as RFC3339").Short('E').
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_END").
		Default(now.Format(timeFmt)).String()

	outStd = kingpin.Flag("out-stdout", "Write output to stdout. Used by default if no other output is specified").
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_STDOUT").
		Default("false").Bool()

	outCSV = kingpin.Flag("out-csv", "Write output to CSV file").
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_CSV").
		Default("false").Bool()
	outCSVFile = kingpin.Flag("out-csv-file", "Set the file to write to").
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_CSV_FILE").
			Default("out.csv").String()

	outJSON = kingpin.Flag("out-json", "Write output to JSON file").
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_JSON").
		Default("false").Bool()
	outJSONFile = kingpin.Flag("out-json-file", "Set the file to write to").
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_JSON_FILE").
			Default("out.json").String()

	outES = kingpin.Flag("out-es", "Index output to elasticsearch").
		OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_ES").
		Default("false").Bool()
	outESIdx = kingpin.Flag("out-es-index", "Elasticsearch index to use for output").
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_ES_INDEX").
			Default("candlestick").String()
	outESHost = kingpin.Flag("out-es-host", "set the elasticsearch host to write to").
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_ES_HOST").
			Default("localhost").String()
	outESPort = kingpin.Flag("out-es-port", "set the elasticsearch port to write to").
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_OUT_ES_PORT").
			Default("9200").String()
	outESSecure = kingpin.Flag("out-es-secure", "set the elasticsearch requests to use https").
			OverrideDefaultFromEnvar("GDAX_EXTRACTOR_SECURE").
			Default("false").Bool()
)

func main() {
	kingpin.Version("1.0.0")
	kingpin.Parse()
	if *verbose {
		printVars()
	}

	xtrct := extractor.New(&extractor.ExtractorConfig{
		Key:        *key,
		Secret:     *secret,
		Passphrase: *passphrase,
		BufferSize: *bufferSize,
		Logger:     Log{},
		Extraction: &extractor.ExtractionConfig{
			Product:     *product,
			Start:       parseTime(*start),
			End:         parseTime(*end),
			Granularity: *granularity,
		},
	})

	fmt.Print("\nExtracting...\n\n")
	started := time.Now()

	err := xtrct.Start()
	check(err)

	collector := extractor.NewCollector(&extractor.CollectorConfig{
		Extractor: xtrct,
	})

	// Write out to a CSV file
	if *outCSV {
		rcv, err := receivers.NewCSV(*outCSVFile)
		check(err)
		collector.Add(rcv)
	}

	// Write out to a JSON file
	if *outJSON {
		rcv, err := receivers.NewJSON(*outJSONFile)
		check(err)
		collector.Add(rcv)
	}

	// Index to elasticsearch
	if *outES {
		rcv, err := receivers.NewElasticsearch(*outESIdx, *outESHost, *outESPort, *outESSecure)
		check(err)
		collector.Add(rcv)
	}

	// log to stdout if no other receiver is set, or stdout is explicitly set
	if *outStd || len(collector.Receivers) == 0 {
		collector.Add(receivers.NewStdout())
	}

	err = collector.Collect()
	check(err)
	fmt.Printf("\n...Done in %s\n", time.Since(started).String())
}

//
// Helpers
//

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func parseTime(date string) time.Time {
	t, err := time.Parse(timeFmt, date)
	if err != nil {
		panic(fmt.Sprintf("Time must be of format [%s]: found [%s]", timeFmt, date))
	}
	return t
}

func printVars() {
	fmt.Printf("Now                     : %s\n", now.Format(timeFmt))
	fmt.Printf("Product ID              : %s\n", *product)
	fmt.Printf("Secret                  : %s\n", *secret)
	fmt.Printf("Key                     : %s\n", *key)
	fmt.Printf("Passphrase              : %s\n", *passphrase)
	fmt.Printf("Granularity             : %d\n", *granularity)
	fmt.Printf("Start                   : %s\n", *start)
	fmt.Printf("End                     : %s\n", *end)

	fmt.Printf("Out stdout              : %t\n", *outStd)

	fmt.Printf("Out CSV                 : %t\n", *outCSV)
	if *outCSV {
		fmt.Printf("Out CSV File            : %s\n", *outCSVFile)
	}

	fmt.Printf("Out JSON                : %t\n", *outJSON)
	if *outJSON {
		fmt.Printf("Out JSON File           : %s\n", *outJSONFile)
	}

	fmt.Printf("Out Elasticsearch       : %t\n", *outES)
	if *outES {
		fmt.Printf("Out Elasticsearch Index : %s\n", *outESIdx)
		fmt.Printf("Out Elasticsearch Host  : %s\n", *outESHost)
		fmt.Printf("Out Elasticsearch Port  : %s\n", *outESPort)
	}

}
